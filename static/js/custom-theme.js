// Custom SQL Playground theme based on the dracula theme
CodeMirror.defineOption("theme", "sql-playground", false);

CodeMirror.defineTheme("sql-playground", {
  base: "dracula",
  inherit: true,
  
  // Override specific styles for SQL keywords
  styles: {
    ".cm-s-sql-playground .cm-keyword": "color: #ff79c6; font-weight: bold",
    ".cm-s-sql-playground .cm-atom": "color: #bd93f9",
    ".cm-s-sql-playground .cm-number": "color: #bd93f9",
    ".cm-s-sql-playground .cm-def": "color: #50fa7b",
    ".cm-s-sql-playground .cm-variable": "color: #f8f8f2",
    ".cm-s-sql-playground .cm-variable-2": "color: #8be9fd",
    ".cm-s-sql-playground .cm-variable-3": "color: #50fa7b",
    ".cm-s-sql-playground .cm-property": "color: #66d9ef",
    ".cm-s-sql-playground .cm-operator": "color: #ff79c6",
    ".cm-s-sql-playground .cm-comment": "color: #6272a4",
    ".cm-s-sql-playground .cm-string": "color: #f1fa8c",
    ".cm-s-sql-playground .cm-string-2": "color: #f1fa8c",
  }
});
