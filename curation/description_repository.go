// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jcodagnone/chapauy/curation/utils"
)

// DescriptionQueueItem represents an item in the description curation queue.
type DescriptionQueueItem struct {
	Description string `json:"description"`
	Count       int    `json:"count"`
}

// Article represents a traffic regulation article.
type Article struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Code  int8   `json:"code"`
	Title string `json:"title"`
}

// Description represents a raw offense description and its classification.
type Description struct {
	ID           int       `json:"id"`
	Description  string    `json:"description"`
	ArticleIDs   []string  `json:"article_ids"`
	ArticleCodes []int8    `json:"article_codes,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReviewDescription represents a description to be reviewed.
type ReviewDescription struct {
	Description  string
	OffenseCount int
}

// ReviewArticle represents an article to be reviewed.
type ReviewArticle struct {
	ID           string
	Text         string
	Descriptions []ReviewDescription
}

// ReviewCode represents a code to be reviewed.
type ReviewCode struct {
	Code     int
	Roman    string
	Articles []ReviewArticle
}

// DescriptionRepository handles the persistence of offense description curations.
type DescriptionRepository interface {
	CreateSchema() error
	SeedArticles(articles []Article) error
	GetUnclassifiedDescriptions(limit int) ([]DescriptionQueueItem, error)
	ListArticles() ([]Article, error)
	ListArticleSections() ([]ValueCount, error)
	SaveDescriptionClassification(description string, articleIDs []string) error
	GetDescriptionProgress() (totalDescriptions, classifiedDescriptions, totalOffenses, classifiedOffenses int, err error)
	// New methods for bulk operations
	GetAllDescriptionJudgmentsSorted() ([]*Description, error)
	BulkInsertDescriptionJudgments(judgments []*Description) error
	CountDescriptionJudgments() (int, error)
	AddArticle(id, text string, code int8, title string) error
	SearchArticles(query string) ([]Article, error)
	CountArticles() (int, error)
	IsDescriptionClassified(description string) (bool, error)
	AreMultiArticlePartsClassified(description string) (bool, error)
	GetDescriptionWithArticles(description string) (*Description, error)
	GetReviewAssignments() ([]ReviewCode, error)
}

type sqlDescriptionRepository struct {
	db *sql.DB
}

// NewDescriptionRepository creates a new description repository.
func NewDescriptionRepository(db *sql.DB) DescriptionRepository {
	return &sqlDescriptionRepository{db: db}
}

func (r *sqlDescriptionRepository) CreateSchema() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id VARCHAR PRIMARY KEY,
			text VARCHAR NOT NULL,
			code TINYINT NOT NULL,
			title VARCHAR NOT NULL
		);

		CREATE SEQUENCE IF NOT EXISTS descriptions_seq;
		CREATE TABLE IF NOT EXISTS descriptions (
			id INTEGER PRIMARY KEY DEFAULT nextval('descriptions_seq'),
			description VARCHAR UNIQUE NOT NULL,
			article_ids VARCHAR[],
			article_codes TINYINT[],
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)

	return err
}

func (r *sqlDescriptionRepository) SeedArticles(articles []Article) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO articles (id, text, code, title) VALUES (?, ?, ?, ?)")
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}

		return err
	}
	defer stmt.Close()

	for _, article := range articles {
		_, err := stmt.Exec(article.ID, article.Text, article.Code, article.Title)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	return tx.Commit()
}

func (r *sqlDescriptionRepository) GetUnclassifiedDescriptions(limit int) ([]DescriptionQueueItem, error) {
	query := `
		SELECT
			o.description,
			COUNT(*) as count
		FROM offenses o
		LEFT JOIN descriptions d ON o.description = d.description
		WHERE o.description IS NOT NULL AND d.description IS NULL
		GROUP BY o.description
		ORDER BY count DESC, o.description ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var descriptions []DescriptionQueueItem

	for rows.Next() {
		var item DescriptionQueueItem
		if err := rows.Scan(&item.Description, &item.Count); err != nil {
			return nil, err
		}

		descriptions = append(descriptions, item)
	}

	return descriptions, nil
}

func (r *sqlDescriptionRepository) ListArticles() ([]Article, error) {
	rows, err := r.db.Query("SELECT id, text, code, title FROM articles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Text, &a.Code, &a.Title); err != nil {
			return nil, err
		}

		articles = append(articles, a)
	}

	return articles, nil
}

func (r *sqlDescriptionRepository) SaveDescriptionClassification(description string, articleIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction saving description classifications for %s: %v", description, err)
		}
	}()

	// 1. Fetch article codes for the given article IDs
	var articleCodes []int8

	if len(articleIDs) > 0 {
		idToCode := make(map[string]int8)

		q := fmt.Sprintf("SELECT id, code FROM articles WHERE id IN (%s)", strings.Repeat("?,", len(articleIDs)-1)+"?") // #nosec G201 - es una buena causa

		args := make([]any, len(articleIDs))
		for i, id := range articleIDs {
			args[i] = id
		}

		rows, err := tx.Query(q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string

			var code int8
			if err := rows.Scan(&id, &code); err != nil {
				return err
			}

			idToCode[id] = code
		}

		uniqueCodes := make(map[int8]bool)

		for _, id := range articleIDs {
			code, ok := idToCode[id]
			if !ok {
				return fmt.Errorf("unknown article ID: %s", id)
			}

			if !uniqueCodes[code] {
				articleCodes = append(articleCodes, code)
				uniqueCodes[code] = true
			}
		}
	}

	// 2. Save to descriptions table
	now := time.Now()

	_, err = tx.Exec(`
		INSERT INTO descriptions (description, article_ids, article_codes, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(description) DO UPDATE SET
			article_ids = excluded.article_ids,
			article_codes = excluded.article_codes,
			updated_at = excluded.updated_at;
	`, description, articleIDs, articleCodes, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAllDescriptionJudgmentsSorted retrieves all description judgments from the database.
func (r *sqlDescriptionRepository) GetAllDescriptionJudgmentsSorted() ([]*Description, error) {
	rows, err := r.db.Query("SELECT description, article_ids, article_codes, updated_at FROM descriptions ORDER BY description")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var judgments []*Description

	for rows.Next() {
		var j Description

		var articleIDs, articleCodes any
		if err := rows.Scan(&j.Description, &articleIDs, &articleCodes, &j.UpdatedAt); err != nil {
			return nil, err
		}

		var ok bool

		j.ArticleIDs, ok = utils.AnyToStringSlice(articleIDs)
		if !ok {
			return nil, fmt.Errorf("failed to convert article_ids to []string for description: %s", j.Description)
		}

		j.ArticleCodes, ok = utils.AnyToInt8Slice(articleCodes)
		if !ok {
			return nil, fmt.Errorf("failed to convert article_codes to []int8 for description: %s", j.Description)
		}

		judgments = append(judgments, &j)
	}

	return judgments, nil
}

// BulkInsertDescriptionJudgments inserts a slice of DescriptionJudgment into the database in bulk.
func (r *sqlDescriptionRepository) BulkInsertDescriptionJudgments(judgments []*Description) error {
	allArticles, err := r.ListArticles()
	if err != nil {
		return err
	}

	idToCode := make(map[string]int8)
	for _, article := range allArticles {
		idToCode[article.ID] = article.Code
	}

	now := time.Now()

	for _, j := range judgments {
		uniqueCodes := make(map[int8]bool)

		var codes []int8

		for _, id := range j.ArticleIDs {
			code, ok := idToCode[id]
			if !ok {
				return fmt.Errorf("unknown article ID: %s", id)
			}

			if !uniqueCodes[code] {
				codes = append(codes, code)
				uniqueCodes[code] = true
			}
		}

		j.ArticleCodes = codes
		if j.UpdatedAt.IsZero() {
			j.UpdatedAt = now
		}
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO descriptions (description, article_ids, article_codes, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(description) DO UPDATE SET
			article_ids = excluded.article_ids,
			article_codes = excluded.article_codes,
			updated_at = excluded.updated_at;
	`)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}

		return err
	}
	defer stmt.Close()

	for _, j := range judgments {
		if _, err := stmt.Exec(j.Description, j.ArticleIDs, j.ArticleCodes, j.UpdatedAt); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}

			return err
		}
	}

	return tx.Commit()
}

// CountDescriptionJudgments counts the number of description judgments in the database.
func (r *sqlDescriptionRepository) CountDescriptionJudgments() (int, error) {
	var count int

	err := r.db.QueryRow("SELECT COUNT(*) FROM descriptions").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetDescriptionProgress calculates the total and classified description counts.
func (r *sqlDescriptionRepository) GetDescriptionProgress() (totalDescriptions, classifiedDescriptions, totalOffenses, classifiedOffenses int, err error) {
	// Total unique descriptions
	queryTotal := `
		SELECT COUNT(DISTINCT description)
		FROM offenses`

	err = r.db.QueryRow(queryTotal).Scan(&totalDescriptions)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Classified descriptions
	queryClassified := `
		SELECT COUNT(DISTINCT description)
		FROM descriptions`

	err = r.db.QueryRow(queryClassified).Scan(&classifiedDescriptions)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Total offenses
	queryTotalOffenses := `SELECT COUNT(*) FROM offenses WHERE description IS NOT NULL AND description != ''`

	err = r.db.QueryRow(queryTotalOffenses).Scan(&totalOffenses)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Classified offenses
	queryClassifiedOffenses := `
		SELECT COUNT(*)
		FROM offenses o
		INNER JOIN descriptions d ON o.description = d.description`

	err = r.db.QueryRow(queryClassifiedOffenses).Scan(&classifiedOffenses)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return totalDescriptions, classifiedDescriptions, totalOffenses, classifiedOffenses, nil
}

// AddArticle inserts a new article into the articles table.
func (r *sqlDescriptionRepository) AddArticle(id, text string, code int8, title string) error {
	_, err := r.db.Exec(`
		INSERT INTO articles (id, text, code, title)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			text = excluded.text,
			code = excluded.code,
			title = excluded.title;
	`, id, text, code, title)

	return err
}

// SearchArticles searches for articles by ID or text.
func (r *sqlDescriptionRepository) SearchArticles(query string) ([]Article, error) {
	// Remove ':' from the query before tokenizing
	cleanedQuery := strings.ReplaceAll(query, ":", "")

	tokens := strings.Fields(cleanedQuery)
	if len(tokens) == 0 {
		return r.ListArticles() // Return all articles if query is empty
	}

	whereClauses := make([]string, 0, len(tokens))
	args := make([]any, 0, len(tokens)*3)
	scoreClauses := make([]string, 0, len(tokens))

	for _, token := range tokens {
		likeToken := "%" + token + "%"

		whereClauses = append(whereClauses, "(id LIKE ? OR text LIKE ? OR title LIKE ?)")
		args = append(args, likeToken, likeToken, likeToken)
		scoreClauses = append(scoreClauses, "(CASE WHEN id LIKE ? OR text LIKE ? OR title LIKE ? THEN 1 ELSE 0 END)")
		args = append(args, likeToken, likeToken, likeToken)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, text, code, title
		FROM articles
		WHERE %s
		ORDER BY (%s) DESC, id
	`, // #nosec G201 - no injections. we use `?' for arguments
		strings.Join(whereClauses, " OR "),
		strings.Join(scoreClauses, " + "),
	)

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article

	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Text, &a.Code, &a.Title); err != nil {
			return nil, err
		}

		articles = append(articles, a)
	}

	return articles, nil
}

// CountArticles counts the number of articles in the articles table.
func (r *sqlDescriptionRepository) CountArticles() (int, error) {
	var count int

	err := r.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqlDescriptionRepository) IsDescriptionClassified(description string) (bool, error) {
	var count int

	err := r.db.QueryRow("SELECT COUNT(*) FROM descriptions WHERE description = ?", description).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AreMultiArticlePartsClassified checks if all comma-separated parts of a multi-article description
// are already classified in the database. Returns true if all parts are classified, false if at least one part is not.
func (r *sqlDescriptionRepository) AreMultiArticlePartsClassified(description string) (bool, error) {
	if !strings.Contains(description, ",") {
		// Not a multi-article description, check the whole thing
		return r.IsDescriptionClassified(description)
	}

	_, found, err := utils.ResolveMultiArticle(
		description,
		func(part string) (utils.Classification, bool, error) {
			classified, err := r.IsDescriptionClassified(part)
			if err != nil {
				return utils.Classification{}, false, err
			}
			// We don't need the actual classification data here, just existence
			return utils.Classification{}, classified, nil
		},
	)

	return found, err
}

// GetDescriptionWithArticles retrieves a description and its associated article IDs from the database.
// Returns nil if the description is not found.
func (r *sqlDescriptionRepository) GetDescriptionWithArticles(description string) (*Description, error) {
	var d Description

	var articleIDs, articleCodes any

	err := r.db.QueryRow("SELECT description, article_ids, article_codes, updated_at FROM descriptions WHERE description = ?", description).Scan(&d.Description, &articleIDs, &articleCodes, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var ok bool

	d.ArticleIDs, ok = utils.AnyToStringSlice(articleIDs)
	if !ok {
		return nil, fmt.Errorf("failed to convert article_ids to []string for description: %s", d.Description)
	}

	d.ArticleCodes, ok = utils.AnyToInt8Slice(articleCodes)
	if !ok {
		return nil, fmt.Errorf("failed to convert article_codes to []int8 for description: %s", d.Description)
	}

	return &d, nil
}

func (r *sqlDescriptionRepository) ListArticleSections() ([]ValueCount, error) {
	rows, err := r.db.Query(`
		SELECT title, COUNT(*) as count
		FROM articles
		GROUP BY title
		ORDER BY count DESC, title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []ValueCount

	for rows.Next() {
		var section ValueCount
		if err := rows.Scan(&section.Value, &section.Count); err != nil {
			return nil, err
		}

		sections = append(sections, section)
	}

	return sections, nil
}

func (r *sqlDescriptionRepository) GetReviewAssignments() ([]ReviewCode, error) {
	query := `
		WITH description_offense_counts AS (
			SELECT
				description,
				COUNT(*) as offense_count
			FROM offenses
			GROUP BY description
		)
						SELECT
							a.code,
							a.id,
							a.text,
							d.description,
							doc.offense_count
						FROM articles a
						LEFT JOIN descriptions d ON list_contains(d.article_ids, a.id) AND len(d.article_ids) = 1
						LEFT JOIN description_offense_counts doc ON d.description = doc.description
						ORDER BY a.code ASC, a.id ASC, d.description ASC;
					`
	rows, err := r.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var reviewCodes []ReviewCode

	codeMap := make(map[int]*ReviewCode)

	for rows.Next() {
		var code int

		var articleID string

		var articleText string

		var description sql.NullString

		var offenseCount sql.NullInt64

		if err := rows.Scan(&code, &articleID, &articleText, &description, &offenseCount); err != nil {
			return nil, err
		}

		if _, ok := codeMap[code]; !ok {
			reviewCodes = append(reviewCodes, ReviewCode{
				Code:  code,
				Roman: utils.ToRoman(code),
			})
			codeMap[code] = &reviewCodes[len(reviewCodes)-1]
		}

		currentCode := codeMap[code]

		var currentArticle *ReviewArticle

		for i := range currentCode.Articles {
			if currentCode.Articles[i].ID == articleID {
				currentArticle = &currentCode.Articles[i]

				break
			}
		}

		if currentArticle == nil {
			currentCode.Articles = append(currentCode.Articles, ReviewArticle{
				ID:   articleID,
				Text: articleText,
			})
			currentArticle = &currentCode.Articles[len(currentCode.Articles)-1]
		}

		// Only add description if it's not NULL
		if description.Valid {
			currentArticle.Descriptions = append(currentArticle.Descriptions, ReviewDescription{
				Description:  description.String,
				OffenseCount: int(offenseCount.Int64),
			})
		}
	}

	return reviewCodes, nil
}

// ValueCount represents a generic value and its count.
type ValueCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
