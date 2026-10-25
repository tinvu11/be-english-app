-- Apply only to an empty, isolated test database after the project migrations.
INSERT INTO users (id,username,email,created_at,updated_at,is_active)
SELECT ('00000000-0000-0000-0000-00000000000'||n)::uuid,
       'dashboard_user_'||n, 'dashboard_'||n||'@example.test',
       (((CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
           - CASE n WHEN 1 THEN 40 WHEN 2 THEN 10 WHEN 3 THEN 2 ELSE 0 END
          + TIME '00:30') AT TIME ZONE 'Asia/Ho_Chi_Minh') AT TIME ZONE 'UTC',
       CURRENT_TIMESTAMP, n<>4
FROM generate_series(1,4) n;
INSERT INTO channels (channel_youtube_id,channel_name,is_active)
VALUES ('dashboard_active','Dashboard channel',true),('dashboard_disabled','Disabled channel',false);
INSERT INTO levels (code,language_id) SELECT 'A1',id FROM languages WHERE code='en';
INSERT INTO topics (slug,is_active) VALUES ('dashboard-active',true),('dashboard-inactive',false);
INSERT INTO videos (title,youtube_id,status,language_id,level_id,channel_id,is_system,created_at)
SELECT 'Dashboard system '||n, 'dashboard_sys_'||n,
       CASE WHEN n<3 THEN 'published' WHEN n<5 THEN 'draft' ELSE 'archived' END,
       l.id,lv.id,c.id,true,
       ((CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh')::date - CASE WHEN n=5 THEN 40 ELSE n END + TIME '01:00') AT TIME ZONE 'Asia/Ho_Chi_Minh'
FROM generate_series(0,5) n, languages l, levels lv, channels c
WHERE l.code='en' AND lv.language_id=l.id AND c.channel_youtube_id='dashboard_active';
INSERT INTO videos (title,youtube_id,status,language_id,level_id,channel_id,is_system)
SELECT 'Private import','dashboard_import','published',l.id,lv.id,c.id,false
FROM languages l,levels lv,channels c
WHERE l.code='en' AND lv.language_id=l.id AND c.channel_youtube_id='dashboard_active';
INSERT INTO video_captions (video_id,sentence_order,start_time_ms,end_time_ms,content)
SELECT v.id,n,n*1000,(n+1)*1000,'Test caption '||n
FROM videos v,generate_series(0,1) n WHERE v.youtube_id='dashboard_sys_0';
INSERT INTO video_captions (video_id,sentence_order,start_time_ms,end_time_ms,content)
SELECT id,0,0,1000,'Private caption' FROM videos WHERE youtube_id='dashboard_import';
INSERT INTO user_watch_history (user_id,video_id,last_watched_at)
SELECT '00000000-0000-0000-0000-000000000003'::uuid,id,CURRENT_TIMESTAMP
FROM videos WHERE youtube_id IN ('dashboard_sys_0','dashboard_sys_1');
INSERT INTO user_watch_history (user_id,video_id,last_watched_at)
SELECT '00000000-0000-0000-0000-000000000002'::uuid,id,CURRENT_TIMESTAMP-INTERVAL '40 days'
FROM videos WHERE youtube_id='dashboard_sys_0';
