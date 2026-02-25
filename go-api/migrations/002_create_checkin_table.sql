CREATE TABLE IF NOT EXISTS qingka_wangke_checkin (
  id INT AUTO_INCREMENT PRIMARY KEY,
  uid INT NOT NULL,
  username VARCHAR(100) NOT NULL DEFAULT '',
  reward_money DECIMAL(10,2) NOT NULL DEFAULT 0,
  checkin_date DATE NOT NULL,
  addtime DATETIME NOT NULL,
  UNIQUE KEY uk_uid_date (uid, checkin_date),
  KEY idx_date (checkin_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
