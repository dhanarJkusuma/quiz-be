CREATE TABLE IF NOT EXISTS quiz (
	id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
	question TEXT NOT NULL,
	active TINYINT NOT NULL DEFAULT 1,

	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS answer (
    id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
    quiz_id INT UNSIGNED NOT NULL,
    answer TEXT NOT NULL,
    correct_answer TINYINT NOT NULL DEFAULT 0,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (quiz_id) REFERENCES quiz(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_history (
    id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
    user_id_p1 INT UNSIGNED NOT NULL,
    user_id_p2 INT UNSIGNED NOT NULL,

    score_p1 INT UNSIGNED NOT NULL DEFAULT 0,
    score_p2 INT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX `user_history_user_id_p1_idx` ON user_history(user_id_p1);
CREATE INDEX `user_history_user_id_p2_idx` ON user_history(user_id_p2);

CREATE TABLE IF NOT EXISTS txn_quiz (
    id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX `txn_quiz_user_id_idx` ON txn_quiz(user_id, start_time);