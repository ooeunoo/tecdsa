CREATE TABLE IF NOT EXISTS request_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(255),
    region VARCHAR(50),
    params TEXT,
    body TEXT,
    method VARCHAR(10),
    headers TEXT,
    ip_address VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);