/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    const dbSelect = document.getElementById('dbSelect');
    const resultsContainer = document.getElementById('results');

    let searchTimeout;

    // Fetch databases and populate the dropdown
    fetch('/api/v1/summary')
        .then(response => response.json())
        .then(data => {
            console.log(data);
            const allOption = document.createElement('option');
            allOption.value = '';
            allOption.textContent = 'All Repositories';
            dbSelect.appendChild(allOption);

            data.databases.forEach(db => {
                const option = document.createElement('option');
                option.value = db.value;
                option.textContent = db.label;
                dbSelect.appendChild(option);
            });
            // Set initial state from URL after databases are loaded
            loadStateFromURL();
        })
        .catch(error => {
            resultsContainer.textContent = 'Error loading databases.';
            console.error('Error:', error);
        });

    // Function to perform the search
    const performSearch = () => {
        const query = searchInput.value.trim();
        const db = dbSelect.value;

        if (!query && !db) {
            resultsContainer.innerHTML = '';
            updateURL('', '');
            return;
        }

        let url = `/api/v1/offenses?`;
        const params = new URLSearchParams();

        if (query) {
            // A simple query parser: "key:value" or just "value"
            if (query.includes(':')) {
                const [key, ...valueParts] = query.split(':');
                const value = valueParts.join(':').trim();
                params.append(key, value);
            } else {
                params.append('description', query);
            }
        }

        if (db) {
            params.append('database', db);
        }

        url += params.toString();

        updateURL(query, db);
        resultsContainer.textContent = 'Searching...';

        fetch(url)
            .then(response => response.json())
            .then(data => {
                renderResults(data);
            })
            .catch(error => {
                resultsContainer.textContent = 'Error performing search.';
                console.error('Error:', error);
            });
    };

    // Function to render results
    const renderResults = (data) => {
        const { offenses, summary } = data;
        if (!offenses || offenses.length === 0) {
            resultsContainer.textContent = 'No results found.';
            return;
        }

        let html = `Total UR: ${summary.total_ur}, Average UR: ${summary.avg_ur}\n\n`;
        offenses.forEach(offense => {
            const time = new Date(offense.time).toLocaleString();
            html += `${time} | ${offense.location} | ${offense.description} | Vehicle: ${offense.vehicle}\n`;
        });
        resultsContainer.textContent = html;
    };

    // Function to update URL query params
    const updateURL = (query, db) => {
        const params = new URLSearchParams();
        if (query) {
            params.set('q', query);
        }
        if (db) {
            params.set('db', db);
        }
        const newUrl = `${window.location.pathname}?${params.toString()}`;
        window.history.pushState({ path: newUrl }, '', newUrl);
    };

    // Function to load state from URL
    const loadStateFromURL = () => {
        const params = new URLSearchParams(window.location.search);
        const query = params.get('q');
        const db = params.get('db');

        if (query) {
            searchInput.value = query;
        }
        if (db) {
            dbSelect.value = db;
        }

        if (query || db) {
            performSearch();
        }
    };


    // Event listeners
    searchInput.addEventListener('input', () => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(performSearch, 300); // Debounce search
    });

    dbSelect.addEventListener('change', performSearch);
});