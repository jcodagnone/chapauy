// Copyright 2025 The ChapaUY Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcodagnone/chapauy/curation"
	"github.com/spf13/cobra"
)

var (
	threshold    float64
	interactive  bool
	multiArticle bool
)

var curationDescriptionCmd = &cobra.Command{
	Use:   "description",
	Short: "Interactive batch curation for descriptions",
	RunE: func(_ *cobra.Command, _ []string) error {
		dbpath := filepath.Join(impoOptions.DbPath, "chapauy.duckdb")
		db, err := sql.Open("duckdb", dbpath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		descrRepo := curation.NewDescriptionRepository(db)
		articles, err := descrRepo.ListArticles()
		if err != nil {
			return fmt.Errorf("listing articles: %w", err)
		}

		descriptions, err := descrRepo.GetAllDescriptionJudgmentsSorted()
		if err != nil {
			return fmt.Errorf("loading classified descriptions: %w", err)
		}

		classifier := curation.NewDescriptionClassifierWithDescriptions(articles, descriptions)

		if interactive {
			// Interactive mode
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Println("Entering interactive mode. Type 'exit' or 'quit' to stop.")
			for {
				fmt.Print("> ")
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()
				if line == "exit" || line == "quit" {
					break
				}

				isMulti := classifier.DetectMultiArticle(line, threshold)

				// Apply multi-article filter if set
				if multiArticle && !isMulti {
					continue
				}
				if !multiArticle && isMulti {
					continue
				}

				if isMulti {
					// Display multi-article description with breakdown
					fmt.Printf("# MULTI | %s\n", line)
					breakdown := classifier.SuggestWithBreakdown(line, threshold)
					for _, bd := range breakdown {
						fmt.Printf("## %s\n", bd.Part)
						for _, suggestion := range bd.Suggestions {
							fmt.Printf("%.2f | %s | %s\n", suggestion.Score, suggestion.ArticleID, suggestion.Text)
						}
					}
				} else {
					// Display single-article description
					suggestions := classifier.Suggest(line, threshold)
					if len(suggestions) > 0 {
						fmt.Printf("# %5s%s\n", "", line)
						for _, suggestion := range suggestions {
							fmt.Printf("%.2f | %4s | %s\n", suggestion.Score, suggestion.ArticleID, suggestion.Text)
						}
					} else {
						fmt.Println("No suggestions found.")
					}
				}
				fmt.Println() // Add a blank line for readability
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
		} else if isTerminal(os.Stdin) {
			// Generation mode
			unclassified, err := descrRepo.GetUnclassifiedDescriptions(10000)
			if err != nil {
				return fmt.Errorf("getting unclassified descriptions: %w", err)
			}

			for _, item := range unclassified {
				isMulti := classifier.DetectMultiArticle(item.Description, threshold)

				// Apply multi-article filter if set
				if multiArticle && !isMulti {
					continue
				}
				if !multiArticle && isMulti {
					continue
				}

				// For multi-article descriptions, check if all parts are already classified
				if isMulti {
					allPartsClassified, err := descrRepo.AreMultiArticlePartsClassified(item.Description)
					if err != nil {
						return fmt.Errorf("checking multi-article parts classification: %w", err)
					}
					if allPartsClassified {
						// Skip if all parts are already classified
						continue
					}

					// Display multi-article description with breakdown
					fmt.Printf("# MULTI | %s\n", item.Description)
					breakdown := classifier.SuggestWithBreakdown(item.Description, threshold)
					for _, bd := range breakdown {
						fmt.Printf("## %s\n", bd.Part)
						for _, suggestion := range bd.Suggestions {
							fmt.Printf("%.2f | %s | %s\n", suggestion.Score, suggestion.ArticleID, suggestion.Text)
						}
					}
				} else {
					// Display single-article description
					suggestions := classifier.Suggest(item.Description, threshold)
					if len(suggestions) > 0 {
						fmt.Printf("# %5s%s\n", "", item.Description)
						for _, suggestion := range suggestions {
							fmt.Printf("%.2f | %s | %s\n", suggestion.Score, suggestion.ArticleID, suggestion.Text)
						}
					}
				}
				fmt.Println() // Add a blank line for readability
			}
		} else {
			// Ingestion mode
			scanner := bufio.NewScanner(os.Stdin)
			var currentDescription string
			var articleIDs []string
			var isMultiDescription bool

			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "# MULTI | ") {
					// Save previous entry if we were processing one
					if currentDescription != "" && len(articleIDs) > 0 {
						classified, err := descrRepo.IsDescriptionClassified(currentDescription)
						if err != nil {
							fmt.Printf("Error checking classification for '%s': %v\n", currentDescription, err)
						} else if classified {
							fmt.Printf("Skipping already classified description: '%s'\n", currentDescription)
						} else {
							if err := descrRepo.SaveDescriptionClassification(currentDescription, articleIDs); err != nil {
								fmt.Printf("Error saving classification for '%s': %v\n", currentDescription, err)
							} else {
								fmt.Printf("Saved classification for '%s'\n", currentDescription)
							}
						}
					}
					// Start new multi-article description
					currentDescription = strings.TrimSpace(strings.TrimPrefix(line, "# MULTI | "))
					articleIDs = []string{}
					isMultiDescription = true
				} else if strings.HasPrefix(line, "# ") {
					// Save previous description if we were processing one
					if currentDescription != "" && len(articleIDs) > 0 {
						classified, err := descrRepo.IsDescriptionClassified(currentDescription)
						if err != nil {
							fmt.Printf("Error checking classification for '%s': %v\n", currentDescription, err)
						} else if classified {
							fmt.Printf("Skipping already classified description: '%s'\n", currentDescription)
						} else {
							if err := descrRepo.SaveDescriptionClassification(currentDescription, articleIDs); err != nil {
								fmt.Printf("Error saving classification for '%s': %v\n", currentDescription, err)
							} else {
								fmt.Printf("Saved classification for '%s'\n", currentDescription)
							}
						}
					}
					// Start new single-article description
					currentDescription = strings.TrimSpace(strings.TrimPrefix(line, "# "))
					articleIDs = []string{}
					isMultiDescription = false
				} else if strings.HasPrefix(line, "## ") {
					// Part marker in multi-article description - update current description
					if isMultiDescription {
						// Save previous part if we have article IDs
						if currentDescription != "" && len(articleIDs) > 0 {
							classified, err := descrRepo.IsDescriptionClassified(currentDescription)
							if err != nil {
								fmt.Printf("Error checking classification for '%s': %v\n", currentDescription, err)
							} else if classified {
								fmt.Printf("Skipping already classified description: '%s'\n", currentDescription)
							} else {
								if err := descrRepo.SaveDescriptionClassification(currentDescription, articleIDs); err != nil {
									fmt.Printf("Error saving classification for '%s': %v\n", currentDescription, err)
								} else {
									fmt.Printf("Saved classification for '%s'\n", currentDescription)
								}
							}
						}
						// Start new part
						currentDescription = strings.TrimSpace(strings.TrimPrefix(line, "## "))
						articleIDs = []string{}
					}
				} else if line != "" && !strings.HasPrefix(line, "#") && currentDescription != "" {
					parts := strings.SplitN(line, " | ", 3)
					if len(parts) == 3 {
						articleID := strings.TrimSpace(parts[1])
						if articleID != "" {
							articleIDs = append(articleIDs, articleID)
						}
					}
				}
			}
			// Save the last description
			if currentDescription != "" && len(articleIDs) > 0 {
				classified, err := descrRepo.IsDescriptionClassified(currentDescription)
				if err != nil {
					fmt.Printf("Error checking classification for '%s': %v\n", currentDescription, err)
				} else if classified {
					fmt.Printf("Skipping already classified description: '%s'\n", currentDescription)
				} else {
					if err := descrRepo.SaveDescriptionClassification(currentDescription, articleIDs); err != nil {
						fmt.Printf("Error saving classification for '%s': %v\n", currentDescription, err)
					} else {
						fmt.Printf("Saved classification for '%s'\n", currentDescription)
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
		}

		return nil
	},
}

func init() {
	curationDescriptionCmd.Flags().Float64Var(&threshold, "threshold", 0.5, "Minimum similarity score to consider a suggestion valid")
	curationDescriptionCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive mode")
	curationDescriptionCmd.Flags().BoolVar(&multiArticle, "multi", false, "Filter to show only descriptions with multiple articles")
	curationCmd.AddCommand(curationDescriptionCmd)
}
