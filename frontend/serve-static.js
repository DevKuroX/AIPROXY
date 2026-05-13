const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 1433;
const OUT_DIR = path.join(__dirname, 'out');

const MIME_TYPES = {
  '.html': 'text/html',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpg',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.txt': 'text/plain'
};

const server = http.createServer((req, res) => {
  console.log('Request:', req.url);
  
  let filePath = path.join(OUT_DIR, req.url);
  
  // Remove query string
  const urlPath = req.url.split('?')[0];
  
  // If no extension, try .html
  if (!path.extname(urlPath)) {
    filePath = path.join(OUT_DIR, urlPath + '.html');
  } else {
    filePath = path.join(OUT_DIR, urlPath);
  }
  
  // Root path
  if (urlPath === '/') {
    filePath = path.join(OUT_DIR, 'index.html');
  }
  
  const extname = path.extname(filePath);
  const contentType = MIME_TYPES[extname] || 'application/octet-stream';
  
  console.log('Serving:', filePath);
  
  fs.readFile(filePath, (error, content) => {
    if (error) {
      if (error.code === 'ENOENT') {
        console.log('File not found:', filePath);
        fs.readFile(path.join(OUT_DIR, 'index.html'), (err, indexContent) => {
          if (err) {
            res.writeHead(404, { 'Content-Type': 'text/html' });
            res.end('404 Not Found', 'utf-8');
          } else {
            res.writeHead(200, { 'Content-Type': 'text/html' });
            res.end(indexContent, 'utf-8');
          }
        });
      } else {
        res.writeHead(500);
        res.end('Server Error: ' + error.code);
      }
    } else {
      res.writeHead(200, { 'Content-Type': contentType });
      res.end(content, 'utf-8');
    }
  });
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}/`);
});
