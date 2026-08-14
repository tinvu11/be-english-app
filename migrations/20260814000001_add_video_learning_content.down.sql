DROP TRIGGER IF EXISTS trg_invalidate_video_learning_content ON video_captions;
DROP FUNCTION IF EXISTS invalidate_video_learning_content();
DROP TABLE IF EXISTS video_dictionary_entries;
DROP TABLE IF EXISTS video_quiz_options;
DROP TABLE IF EXISTS video_quizzes;
DROP TABLE IF EXISTS video_summaries;
