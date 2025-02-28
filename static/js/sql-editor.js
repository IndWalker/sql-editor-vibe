document.addEventListener('DOMContentLoaded', function() {
    // State management
    const state = {
        editor: null,
        selectedDialect: 'sqlite',
        darkMode: localStorage.getItem('darkMode') === 'true',
        executeInProgress: false,
        lastResults: null,
        dbStatuses: {
            sqlite: false,
            mysql: false,
            postgresql: false
        },
        sortState: {
            column: null,
            direction: 'asc'
        }
    };

    // Sample queries for each dialect
    const sampleQueries = {
        sqlite: [
            { 
                name: 'Select all data',
                description: 'Retrieve all records from the test_data table',
                query: 'SELECT * FROM test_data LIMIT 10;'
            },
            {
                name: 'Filter by value',
                description: 'Find records with value greater than 300',
                query: 'SELECT * FROM test_data WHERE value > 300;'
            },
            {
                name: 'Sum of values',
                description: 'Calculate the total sum of all values',
                query: 'SELECT SUM(value) AS total_value FROM test_data;'
            }
        ],
        mysql: [
            {
                name: 'Products by category',
                description: 'List products grouped by category',
                query: 'SELECT category, COUNT(*) as count, AVG(price) as avg_price\nFROM products\nGROUP BY category\nORDER BY count DESC;'
            },
            {
                name: 'Expensive electronics',
                description: 'Find electronics that cost more than $500',
                query: 'SELECT name, price, stock FROM products\nWHERE category = \'Electronics\' AND price > 500\nORDER BY price DESC;'
            },
            {
                name: 'Low stock items',
                description: 'Find products with less than 20 items in stock',
                query: 'SELECT * FROM products WHERE stock < 20\nORDER BY stock ASC;'
            }
        ],
        postgresql: [
            {
                name: 'US customers',
                description: 'List all customers from the United States',
                query: 'SELECT first_name, last_name, city, email\nFROM customers\nWHERE country = \'USA\'\nORDER BY city;'
            },
            {
                name: 'Customer locations',
                description: 'Count customers by country',
                query: 'SELECT country, COUNT(*) as customer_count\nFROM customers\nGROUP BY country\nORDER BY customer_count DESC;'
            },
            {
                name: 'Search by name',
                description: 'Find customers with "son" in their last name',
                query: 'SELECT * FROM customers\nWHERE last_name LIKE \'%son%\';'
            }
        ]
    };

    // Database info for SQL hints
    const databaseSchemas = {
        sqlite: {
            test_data: ["id", "name", "value"]
        },
        mysql: {
            products: ["id", "name", "description", "price", "category", "stock", "created_at"]
        },
        postgresql: {
            customers: ["id", "first_name", "last_name", "email", "phone", "country", "city", "address", "postal_code", "created_at"]
        }
    };

    // Icons for each dialect
    const dialectIcons = {
        sqlite: '<i class="fas fa-file-alt mr-2"></i>',
        mysql: '<i class="fas fa-database mr-2"></i>',
        postgresql: '<i class="fas fa-server mr-2"></i>'
    };

    // DOM Elements
    const elements = {
        body: document.getElementById('body'),
        sqlEditor: document.getElementById('sqlEditor'),
        executeQueryBtn: document.getElementById('executeQueryBtn'),
        formatSqlBtn: document.getElementById('formatSqlBtn'),
        clearEditorBtn: document.getElementById('clearEditorBtn'),
        toggleThemeBtn: document.getElementById('toggleThemeBtn'),
        exportResultsBtn: document.getElementById('exportResultsBtn'),
        copyResultsBtn: document.getElementById('copyResultsBtn'),
        csvExportBtn: document.getElementById('csvExportBtn'),
        dbConnections: document.getElementById('dbConnections'),
        sampleQueries: document.getElementById('sampleQueries'),
        dialectBadge: document.getElementById('dialectBadge'),
        cursorPosition: document.getElementById('cursorPosition'),
        editorStats: document.getElementById('editorStats'),
        queryLoader: document.getElementById('queryLoader'),
        errorContainer: document.getElementById('errorContainer'),
        errorMessage: document.getElementById('errorMessage'),
        resultsContainer: document.getElementById('resultsContainer'),
        resultCount: document.getElementById('resultCount'),
        resultsTableHead: document.getElementById('resultsTableHead'),
        resultsTableBody: document.getElementById('resultsTableBody'),
        rowLimitWarning: document.getElementById('rowLimitWarning'),
        emptyResultsContainer: document.getElementById('emptyResultsContainer'),
        shortcutsModal: document.getElementById('shortcutsModal'),
        showShortcutsBtn: document.getElementById('showShortcutsBtn'),
        closeShortcutsBtn: document.getElementById('closeShortcutsBtn'),
        dialectModal: document.getElementById('dialectModal'),
        dialectOptions: document.getElementById('dialectOptions'),
        toastContainer: document.getElementById('toastContainer')
    };

    // Initialize the editor
    function initializeEditor() {
        // Initialize CodeMirror
        state.editor = CodeMirror.fromTextArea(elements.sqlEditor, {
            mode: 'text/x-sql',
            theme: state.darkMode ? 'dracula' : 'default',
            lineNumbers: true,
            matchBrackets: true,
            indentWithTabs: false,
            indentUnit: 4,
            smartIndent: true,
            lineWrapping: true,
            extraKeys: {
            "Ctrl-Space": "autocomplete",
            "Ctrl-Enter": executeQuery,
            "F1": formatSql
            },
            hintOptions: {
            tables: databaseSchemas[state.selectedDialect]
            },
            styleActiveLine: true,
            autoCloseBrackets: true,
            foldGutter: true,
            gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"]
        });

        // Set default query
        state.editor.setValue(sampleQueries[state.selectedDialect][0].query);

        // Update cursor position and editor stats
        state.editor.on('cursorActivity', function() {
            updateEditorInfo();
        });

        state.editor.on('changes', function() {
            updateEditorInfo();
        });
    }

    // Update editor information (cursor position and character count)
    function updateEditorInfo() {
        const cursor = state.editor.getCursor();
        elements.cursorPosition.textContent = `Line: ${cursor.line + 1}, Column: ${cursor.ch + 1}`;
        
        const content = state.editor.getValue();
        elements.editorStats.textContent = `${content.length} characters`;
    }

    // Execute the SQL query
    function executeQuery() {
        if (state.executeInProgress) return;
        
        state.executeInProgress = true;
        elements.queryLoader.classList.remove('hidden');
        elements.executeQueryBtn.classList.add('opacity-70', 'cursor-not-allowed');
        
        // Reset result containers
        hideError();
        hideResults();
        hideEmptyResults();
        
        // Get SQL query
        const sql = state.editor.getValue();
        
        // Validate and execute the query
        fetch('/api/validate-sql', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                sql: sql,
                dialect: state.selectedDialect
            }),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response was not ok. Status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            if (!data.valid) {
                showError(data.error);
                return;
            }

            if (data.error) {
                showError(data.error);
                return;
            }
            
            // Handle successful query
            if (data.result) {
                state.lastResults = data.result;
                
                if (data.result.columns && data.result.columns.length > 0 && data.result.rows && data.result.rows.length > 0) {
                    displayResults(data.result);
                } else {
                    showEmptyResults();
                }
            } else {
                showEmptyResults();
            }
        })
        .catch(error => {
            showError(`Failed to execute query: ${error.message}`);
        })
        .finally(() => {
            state.executeInProgress = false;
            elements.queryLoader.classList.add('hidden');
            elements.executeQueryBtn.classList.remove('opacity-70', 'cursor-not-allowed');
        });
    }

    // Display query results in the table
    function displayResults(result) {
        // Enable export buttons
        elements.exportResultsBtn.disabled = false;
        elements.exportResultsBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        
        // Update result count
        elements.resultCount.textContent = `${result.rows.length} row${result.rows.length !== 1 ? 's' : ''}`;
        
        // Create table header
        const headerRow = document.createElement('tr');
        headerRow.className = 'bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider dark:bg-gray-700 dark:text-gray-300';
        
        result.columns.forEach((column, index) => {
            const th = document.createElement('th');
            th.className = 'px-6 py-3 border-b border-gray-200 dark:border-gray-600';
            
            // Create a container for column name and sort icon
            const container = document.createElement('div');
            container.className = 'flex items-center';
            container.textContent = column;
            
            // Add sort button
            const sortBtn = document.createElement('button');
            sortBtn.className = 'ml-1 text-gray-400 hover:text-gray-600 focus:outline-none dark:text-gray-500 dark:hover:text-gray-300';
            sortBtn.innerHTML = '<i class="fas fa-sort"></i>';
            sortBtn.onclick = function() {
                sortResultsByColumn(index);
            };
            container.appendChild(sortBtn);
            
            th.appendChild(container);
            headerRow.appendChild(th);
        });
        
        elements.resultsTableHead.innerHTML = '';
        elements.resultsTableHead.appendChild(headerRow);
        
        // Create table body rows
        elements.resultsTableBody.innerHTML = '';
        
        result.rows.forEach((row, rowIndex) => {
            const tr = document.createElement('tr');
            tr.className = rowIndex % 2 === 0 ? 'bg-white dark:bg-gray-800' : 'bg-gray-50 dark:bg-gray-700';
            tr.classList.add('hover:bg-gray-100', 'dark:hover:bg-gray-600', 'transition-colors');
            
            row.forEach(cell => {
                const td = document.createElement('td');
                td.className = 'px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-300';
                
                if (cell === null) {
                    const nullSpan = document.createElement('span');
                    nullSpan.className = 'text-gray-400 italic dark:text-gray-500';
                    nullSpan.textContent = 'NULL';
                    td.appendChild(nullSpan);
                } else if (typeof cell === 'number') {
                    const numSpan = document.createElement('span');
                    numSpan.className = 'font-mono text-blue-600 dark:text-blue-400';
                    numSpan.textContent = cell;
                    td.appendChild(numSpan);
                } else {
                    td.textContent = cell;
                }
                
                tr.appendChild(td);
            });
            
            elements.resultsTableBody.appendChild(tr);
        });
        
        // Show the results container
        elements.resultsContainer.classList.remove('hidden');
        
        // Show row limit warning if needed
        if (result.rows.length >= 10) {
            elements.rowLimitWarning.classList.remove('hidden');
        } else {
            elements.rowLimitWarning.classList.add('hidden');
        }
    }

    // Sort results by column
    function sortResultsByColumn(columnIndex) {
        if (!state.lastResults || !state.lastResults.rows || state.lastResults.rows.length === 0) {
            return;
        }
        
        // Toggle sort direction if same column is clicked again
        if (state.sortState.column === columnIndex) {
            state.sortState.direction = state.sortState.direction === 'asc' ? 'desc' : 'asc';
        } else {
            state.sortState.column = columnIndex;
            state.sortState.direction = 'asc';
        }
        
        // Sort the rows
        const sortedRows = [...state.lastResults.rows].sort((a, b) => {
            const valueA = a[columnIndex];
            const valueB = b[columnIndex];
            
            // Handle NULL values
            if (valueA === null && valueB === null) return 0;
            if (valueA === null) return state.sortState.direction === 'asc' ? -1 : 1;
            if (valueB === null) return state.sortState.direction === 'asc' ? 1 : -1;
            
            // Sort based on data type
            if (typeof valueA === 'number' && typeof valueB === 'number') {
                return state.sortState.direction === 'asc' 
                    ? valueA - valueB 
                    : valueB - valueA;
            }
            
            // Default string comparison
            const stringA = String(valueA).toLowerCase();
            const stringB = String(valueB).toLowerCase();
            
            if (state.sortState.direction === 'asc') {
                return stringA.localeCompare(stringB);
            } else {
                return stringB.localeCompare(stringA);
            }
        });
        
        // Update results with sorted rows
        const sortedResult = {
            columns: state.lastResults.columns,
            rows: sortedRows
        };
        
        displayResults(sortedResult);
    }

    // Format SQL query
    function formatSql() {
        const sqlQuery = state.editor.getValue();
        if (!sqlQuery.trim()) return;
        
        try {
            // Simple SQL formatting - capitalize SQL keywords
            const keywords = [
                'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'GROUP BY', 'ORDER BY', 
                'HAVING', 'LIMIT', 'OFFSET', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN', 
                'INNER JOIN', 'OUTER JOIN', 'ON', 'AS', 'UNION', 'ALL', 'INSERT INTO',
                'VALUES', 'UPDATE', 'SET', 'DELETE', 'CREATE', 'ALTER', 'DROP', 'TABLE',
                'INDEX', 'VIEW', 'FUNCTION', 'PROCEDURE', 'TRIGGER', 'CASE', 'WHEN',
                'THEN', 'ELSE', 'END', 'IS', 'NOT', 'NULL', 'IN', 'EXISTS', 'DISTINCT'
            ];
            
            let formattedQuery = sqlQuery;
            
            // Capitalize keywords
            keywords.forEach(keyword => {
                const regex = new RegExp(`\\b${keyword}\\b`, 'gi');
                formattedQuery = formattedQuery.replace(regex, keyword);
            });
            
            // Add line breaks after common clauses
            const clausesToBreakAfter = [
                'SELECT', 'FROM', 'WHERE', 'GROUP BY', 'ORDER BY', 
                'HAVING', 'LIMIT', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN', 
                'INNER JOIN', 'OUTER JOIN'
            ];
            
            clausesToBreakAfter.forEach(clause => {
                const regex = new RegExp(`${clause}\\s`, 'g');
                formattedQuery = formattedQuery.replace(regex, `${clause}\n  `);
            });
            
            // Replace multiple spaces with single space
            formattedQuery = formattedQuery.replace(/\s{2,}/g, ' ');
            
            // Set the formatted query in the editor
            state.editor.setValue(formattedQuery);
            
            showToast('Success', 'SQL query formatted', 'success');
        } catch (error) {
            showToast('Error', 'Failed to format SQL query', 'error');
        }
    }

    // Clear the editor
    function clearEditor() {
        state.editor.setValue('');
        state.editor.focus();
    }

    // Toggle dark/light theme
    function toggleTheme() {
        state.darkMode = !state.darkMode;
        localStorage.setItem('darkMode', state.darkMode);
        
        if (state.darkMode) {
            document.documentElement.classList.add('dark');
            elements.body.classList.add('dark');
            elements.toggleThemeBtn.innerHTML = '<i class="fas fa-sun"></i>';
            state.editor.setOption('theme', 'dracula');
        } else {
            document.documentElement.classList.remove('dark');
            elements.body.classList.remove('dark');
            elements.toggleThemeBtn.innerHTML = '<i class="fas fa-moon"></i>';
            state.editor.setOption('theme', 'default');
        }
    }

    // Change SQL dialect
    function changeDialect(dialect) {
        if (state.selectedDialect === dialect) return;
        
        state.selectedDialect = dialect;
        elements.dialectBadge.textContent = dialect;
        
        // Update editor hints for the selected dialect
        state.editor.setOption('hintOptions', {
            tables: databaseSchemas[dialect]
        });
        
        // Update sample queries display
        renderSampleQueries();
        
        // Load a default query for the selected dialect
        state.editor.setValue(sampleQueries[dialect][0].query);
        
        // Update visual indicator in the connections list
        updateDatabaseConnectionsList();
    }

    // Show sample query in the editor
    function loadSampleQuery(query) {
        state.editor.setValue(query);
        state.editor.focus();
    }

    // Export results to CSV
    function exportToCSV() {
        if (!state.lastResults || !state.lastResults.columns || !state.lastResults.rows) {
            showToast('Error', 'No results to export', 'error');
            return;
        }
        
        // Create CSV content
        let csvContent = state.lastResults.columns.join(',') + '\n';
        
        state.lastResults.rows.forEach(row => {
            const rowStr = row.map(cell => {
                if (cell === null) return 'NULL';
                if (typeof cell === 'string' && cell.includes(',')) {
                    // Escape quotes and wrap in quotes
                    return `"${cell.replace(/"/g, '""')}"`;
                }
                return cell;
            }).join(',');
            csvContent += rowStr + '\n';
        });
        
        // Create download link
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.setAttribute('href', url);
        link.setAttribute('download', `query_results_${formatDate(new Date())}.csv`);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        showToast('Success', 'Results exported to CSV', 'success');
    }

    // Copy results to clipboard
    function copyResultsToClipboard() {
        if (!state.lastResults || !state.lastResults.columns || !state.lastResults.rows) {
            showToast('Error', 'No results to copy', 'error');
            return;
        }
        
        // Create a plain text representation of the table
        let textContent = state.lastResults.columns.join('\t') + '\n';
        
        state.lastResults.rows.forEach(row => {
            textContent += row.map(cell => cell === null ? 'NULL' : String(cell)).join('\t') + '\n';
        });
        
        // Copy to clipboard
        navigator.clipboard.writeText(textContent)
            .then(() => {
                showToast('Success', 'Results copied to clipboard', 'success');
            })
            .catch(() => {
                showToast('Error', 'Failed to copy results', 'error');
            });
    }

    // Show error message
    function showError(message) {
        elements.errorMessage.textContent = message;
        elements.errorContainer.classList.remove('hidden');
    }

    // Hide error message
    function hideError() {
        elements.errorContainer.classList.add('hidden');
        elements.errorMessage.textContent = '';
    }

    // Show results container
    function showResults() {
        elements.resultsContainer.classList.remove('hidden');
    }

    // Hide results container
    function hideResults() {
        elements.resultsContainer.classList.add('hidden');
        elements.resultsTableHead.innerHTML = '';
        elements.resultsTableBody.innerHTML = '';
    }

    // Show empty results message
    function showEmptyResults() {
        elements.emptyResultsContainer.classList.remove('hidden');
    }

    // Hide empty results message
    function hideEmptyResults() {
        elements.emptyResultsContainer.classList.add('hidden');
    }

    // Show toast notification
    function showToast(title, message, type = 'info') {
        const toast = document.createElement('div');
        let bgClass, textClass, iconClass;
        
        if (type === 'success') {
            bgClass = 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800';
            textClass = 'text-green-800 dark:text-green-300';
            iconClass = 'text-green-500 dark:text-green-400';
        } else if (type === 'error') {
            bgClass = 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800';
            textClass = 'text-red-800 dark:text-red-300';
            iconClass = 'text-red-500 dark:text-red-400';
        } else {
            bgClass = 'bg-blue-50 border-blue-200 dark:bg-blue-900/20 dark:border-blue-800';
            textClass = 'text-blue-800 dark:text-blue-300';
            iconClass = 'text-blue-500 dark:text-blue-400';
        }
        
        toast.className = `toast ${bgClass} border rounded-lg shadow-lg`;
        
        const icon = type === 'success' ? 'fa-check-circle' : type === 'error' ? 'fa-exclamation-circle' : 'fa-info-circle';
        
        toast.innerHTML = `
            <div class="flex p-4">
                <div class="flex-shrink-0 ${iconClass}">
                    <i class="fas ${icon}"></i>
                </div>
                <div class="ml-3">
                    <p class="font-medium ${textClass}">${title}</p>
                    <p class="text-sm ${textClass} opacity-90">${message}</p>
                </div>
                <div class="ml-auto pl-3">
                    <button class="toast-close text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
        `;
        
        elements.toastContainer.appendChild(toast);
        
        // Show the toast with animation
        setTimeout(() => {
            toast.classList.add('show');
        }, 10);
        
        // Close button
        const closeBtn = toast.querySelector('.toast-close');
        closeBtn.addEventListener('click', () => {
            toast.classList.remove('show');
            setTimeout(() => {
                elements.toastContainer.removeChild(toast);
            }, 300);
        });
        
        // Auto close after 3 seconds
        setTimeout(() => {
            if (elements.toastContainer.contains(toast)) {
                toast.classList.remove('show');
                setTimeout(() => {
                    if (elements.toastContainer.contains(toast)) {
                        elements.toastContainer.removeChild(toast);
                    }
                }, 300);
            }
        }, 3000);
    }

    // Check database connection status
    function checkDatabaseConnections() {
        fetch('/api/db-status')
            .then(response => response.json())
            .then(data => {
                state.dbStatuses = data;
                updateDatabaseConnectionsList();
            })
            .catch(error => {
                console.error('Failed to check database status:', error);
            });
    }

    // Update database connections list
    function updateDatabaseConnectionsList() {
        let html = '';
        
        ['sqlite', 'mysql', 'postgresql'].forEach(dialect => {
            const isActive = state.selectedDialect === dialect;
            const isConnected = state.dbStatuses[dialect];
            
            html += `
                <div 
                    class="flex items-center justify-between p-2 rounded-md cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 ${isActive ? 'bg-primary-50 dark:bg-primary-900/20' : ''}" 
                    data-dialect="${dialect}">
                    <div class="flex items-center">
                        ${dialectIcons[dialect]}
                        <span class="capitalize dark:text-gray-300">${dialect}</span>
                    </div>
                    <div class="flex items-center">
                        <span 
                            class="status-dot ${isConnected ? 'connected' : 'disconnected'}">
                        </span>
                        <span class="text-xs text-gray-500 dark:text-gray-400">
                            ${isConnected ? 'Connected' : 'Offline'}
                        </span>
                    </div>
                </div>
            `;
        });
        
        elements.dbConnections.innerHTML = html;
        
        // Add click event listeners
        document.querySelectorAll('#dbConnections [data-dialect]').forEach(el => {
            el.addEventListener('click', () => {
                changeDialect(el.dataset.dialect);
            });
        });
    }

    // Render sample queries based on selected dialect
    function renderSampleQueries() {
        const queries = sampleQueries[state.selectedDialect];
        let html = '';
        
        queries.forEach(sample => {
            html += `
                <div 
                    class="p-2 rounded-md cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 text-sm sample-query" 
                    data-query="${encodeURIComponent(sample.query)}">
                    <div class="font-medium mb-1 dark:text-gray-300">${sample.name}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">${sample.description}</div>
                </div>
            `;
        });
        
        elements.sampleQueries.innerHTML = html;
        
        // Add click event listeners
        document.querySelectorAll('.sample-query').forEach(el => {
            el.addEventListener('click', () => {
                const query = decodeURIComponent(el.dataset.query);
                loadSampleQuery(query);
            });
        });
    }

    // Helper function to format date for filenames
    function formatDate(date) {
        return date.toISOString().replace(/[:.]/g, '-').split('T')[0];
    }

    // Helper function to check if a value is a number
    function isNumber(value) {
        return typeof value === 'number' || (typeof value === 'string' && !isNaN(value));
    }

    // Show shortcuts modal
    function showShortcutsModal() {
        elements.shortcutsModal.classList.remove('hidden');
    }

    // Hide shortcuts modal
    function hideShortcutsModal() {
        elements.shortcutsModal.classList.add('hidden');
    }

    // Initialize the application
    function init() {
        // Initialize editor
        initializeEditor();
        
        // Apply dark mode if needed
        if (state.darkMode) {
            document.documentElement.classList.add('dark');
            elements.body.classList.add('dark');
            elements.toggleThemeBtn.innerHTML = '<i class="fas fa-sun"></i>';
        }
        
        // Render initial UI
        updateDatabaseConnectionsList();
        renderSampleQueries();
        
        // Check database connections
        checkDatabaseConnections();
        
        // Set up event listeners
        elements.executeQueryBtn.addEventListener('click', executeQuery);
        elements.formatSqlBtn.addEventListener('click', formatSql);
        elements.clearEditorBtn.addEventListener('click', clearEditor);
        elements.toggleThemeBtn.addEventListener('click', toggleTheme);
        elements.exportResultsBtn.addEventListener('click', exportToCSV);
        elements.csvExportBtn.addEventListener('click', exportToCSV);
        elements.copyResultsBtn.addEventListener('click', copyResultsToClipboard);
        elements.showShortcutsBtn.addEventListener('click', showShortcutsModal);
        elements.closeShortcutsBtn.addEventListener('click', hideShortcutsModal);
        
        // Keyboard shortcuts
        document.addEventListener('keydown', function(e) {
            // Ctrl+Enter - Execute query
            if (e.ctrlKey && e.key === 'Enter') {
                executeQuery();
                e.preventDefault();
            }
            
            // F1 - Format SQL
            if (e.key === 'F1') {
                formatSql();
                e.preventDefault();
            }
            
            // Shift+D - Toggle dark mode
            if (e.shiftKey && e.key === 'D') {
                toggleTheme();
                e.preventDefault();
            }
            
            // Escape - Close modals
            if (e.key === 'Escape') {
                hideShortcutsModal();
            }
        });
        
        // Periodically check database connections
        setInterval(checkDatabaseConnections, 30000); // every 30 seconds
    }

    // Start the application
    init();
});
