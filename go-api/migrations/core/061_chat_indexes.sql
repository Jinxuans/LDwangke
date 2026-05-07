-- 聊天会话与未读查询辅助索引。
ALTER TABLE `qingka_chat_list`
  ADD KEY `idx_chat_user_pair` (`user1`, `user2`);

ALTER TABLE `qingka_chat_msg`
  ADD KEY `idx_chat_to_status_list` (`to_uid`, `status`, `list_id`);
