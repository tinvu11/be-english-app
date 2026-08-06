ALTER TABLE video_topics
    DROP CONSTRAINT IF EXISTS video_topics_topic_id_fkey,
    ADD CONSTRAINT video_topics_topic_id_fkey
        FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE RESTRICT;
