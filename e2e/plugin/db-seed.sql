-- Clean up all existing data (except migrations)
delete from commenters;
delete from commentersessions;
delete from comments;
delete from config;
delete from domains;
delete from emails;
delete from exports;
delete from moderators;
delete from ownerconfirmhexes;
delete from owners;
delete from ownersessions;
delete from pages;
delete from resethexes;
delete from ssotokens;
delete from views;
delete from votes;

-- Insert seed test data
insert into owners(ownerhex, email, name, passwordhash, confirmedemail, joindate)
    values
        ('0000000000000000000000000000000000000000000000000000000000000001', 'ace@comentario.app', 'Captain Ace', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu', 'true', '2023-01-17 17:55:47.008851'),
        ('0000000000000000000000000000000000000000000000000000000000000002', 'king@comentario.app', 'Engineer King', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu', 'true', '2023-01-17 17:55:47.008851'),
        ('0000000000000000000000000000000000000000000000000000000000000003', 'queen@comentario.app', 'Cook Queen', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu', 'true', '2023-01-17 17:55:47.008851'),
        ('0000000000000000000000000000000000000000000000000000000000000004', 'jack@comentario.app', 'Navigator Jack', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu', 'true', '2023-01-17 17:55:47.008851');

insert into commenters (commenterhex, email, name, link, photo, provider, joindate, state, passwordhash)
    values
        ('0000000000000000000000000000000000000000000000000000000000001001', 'ace@comentario.app', 'Captain Ace', 'undefined', 'undefined', 'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu'),
        ('0000000000000000000000000000000000000000000000000000000000001002', 'king@comentario.app', 'Engineer King', 'undefined', 'undefined', 'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu'),
        ('0000000000000000000000000000000000000000000000000000000000001003', 'queen@comentario.app', 'Cook Queen', 'undefined', 'undefined', 'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu'),
        ('0000000000000000000000000000000000000000000000000000000000001004', 'jack@comentario.app', 'Navigator Jack', 'undefined', 'undefined', 'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$NRp62h1E765Rh.VqMfvz2OS9EG92v/BReep4NJbVa7PEKYTWAAJPu'),
        ('0000000000000000000000000000000000000000000000000000000000001010', 'one@blog.com', 'Commenter One', 'undefined', 'undefined', 'commento', '2023-01-18 16:52:04.541982', 'ok', '$2a$10$3w4LEMCh1iKwJC2uMGCP0eb0BRULg77KmnZuvnlGBMs4ALDbJ5Syy'),
        ('0000000000000000000000000000000000000000000000000000000000001011', 'two@blog.com', 'Commenter Two', 'https://wikipedia.org/', 'undefined', 'commento', '2023-01-18 16:52:04.541982', 'ok', '$2a$10$3w4LEMCh1iKwJC2uMGCP0eb0BRULg77KmnZuvnlGBMs4ALDbJ5Syy');

insert into domains(domain, ownerhex, name, creationdate, state, importedcomments, autospamfilter,
                    requiremoderation, requireidentification, viewsthismonth, moderateallanonymous,
                    emailnotificationpolicy, commentoprovider, googleprovider, twitterprovider, githubprovider,
                    gitlabprovider, ssoprovider, ssosecret, ssourl, defaultsortpolicy)
    values
        ('localhost:8000', '0000000000000000000000000000000000000000000000000000000000000001', 'Test Domain',
         '2023-01-17 17:56:10.966890', 'unfrozen', 'false', true, false, false, 0, false, 'pending-moderation', true, true,
         true, true, true, false, '', '', 'score-desc');

insert into emails(email, unsubscribesecrethex, lastemailnotificationdate, pendingemails, sendreplynotifications, sendmoderatornotifications)
    values
        ('ace@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcb8', '2023-01-17 17:55:46.953534', 0, false, true),
        ('king@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcb9', '2023-01-17 17:55:46.953534', 0, false, true),
        ('queen@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcba', '2023-01-17 17:55:46.953534', 0, false, true),
        ('jack@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcbb', '2023-01-17 17:55:46.953534', 0, false, true),
        ('one@blog.com', '2690cab8b021140dfb7d6a56ac60ac49cae3e4706a2e90b4b5645584f59451c7', '2023-01-18 16:52:04.448105', 0, false, true),
        ('two@blog.com', '2690cab8b021140dfb7d6a56ac60ac49cae3e4706a2e90b4b5645584f59451c8', '2023-01-18 16:52:04.448105', 0, false, true);

insert into moderators(domain, email, adddate)
    values
        ('localhost:8000', 'root@comentario.app', '2023-01-17 17:56:10.968427');

insert into pages(domain, path, islocked, commentcount, stickycommenthex, title)
    values
        ('localhost:8000', '/', false, 1, 'none', '');

insert into comments(commenthex, domain, path, commenterhex, markdown, html, parenthex, score, state, creationdate, deleted, deleterhex, deletiondate)
    values
        ('0000000000000000000000000000000000000000000000000000000000002001', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Alright crew, let''s gather around for a quick meeting. We''ve got a **long** voyage ahead of us, and I want to make sure everyone is on the same page.', '<p>Alright crew, let''s gather around for a quick meeting. We''ve got a <b>long</b> voyage ahead of us, and I want to make sure everyone is on the same page.</p>', 'root', 4, 'approved', '2023-01-18 16:52:36.161440', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002002', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001002', 'What''s on the agenda, captain?', '<p>What&#39;s on the agenda, captain?</p>', '0000000000000000000000000000000000000000000000000000000000002001', 0, 'approved', '2023-02-27 18:24:22.057000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002003', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'First off, we need to make sure the engine is in good working order. Any issues we need to address, *engineer*?', '<p>First off, we need to make sure the engine is in good working order. Any issues we need to address, <em>engineer</em>?</p>', '0000000000000000000000000000000000000000000000000000000000002002', 0, 'approved', '2023-02-27 18:25:04.104000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002004', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001002', 'Nothing major, captain. Just some routine maintenance to do, but we should be good to go soon.', '<p>Nothing major, captain. Just some routine maintenance to do, but we should be good to go soon.</p>', '0000000000000000000000000000000000000000000000000000000000002003', 0, 'approved', '2023-02-27 18:26:05.636000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002005', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Good work, navigator. That''s what I was thinking too.', '<p>Good work, navigator. That&#39;s what I was thinking too.</p>', '000000000000000000000000000000000000000000000000000000000000200d', 0, 'approved', '2023-02-27 18:29:25.703000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002006', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'What about supplies, cook?', '<p>What about supplies, cook?</p>', '0000000000000000000000000000000000000000000000000000000000002002', 0, 'approved', '2023-02-27 18:29:39.719000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002007', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Absolutely, cook. I''ll make a note of it.', '<p>Absolutely, cook. I&#39;ll make a note of it.</p>', '0000000000000000000000000000000000000000000000000000000000002010', 0, 'approved', '2023-02-27 18:33:19.502000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002008', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Now, is there anything else anyone wants to bring up?', '<p>Now, is there anything else anyone wants to bring up?</p>', 'root', 0, 'approved', '2023-02-27 18:33:24.642000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002009', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001002', 'Captain, I''ve been noticing some strange vibrations in the engine room. It''s nothing too serious, but I''d like to take a look at it just to be safe.', '<p>Captain, I&#39;ve been noticing some strange vibrations in the engine room. It&#39;s nothing too serious, but I&#39;d like to take a look at it just to be safe.</p>', '0000000000000000000000000000000000000000000000000000000000002008', 0, 'approved', '2023-02-27 18:34:15.541000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200a', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Good point, navigator. I''ll make sure our crew is well-armed and that we have extra lookouts posted. Safety is our top priority, after all.', '<p>Good point, navigator. I&#39;ll make sure our crew is well-armed and that we have extra lookouts posted. Safety is our top priority, after all.</p>', '000000000000000000000000000000000000000000000000000000000000200b', 0, 'approved', '2023-02-27 18:35:26.275000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200b', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001004', '**Captain**, one more thing. We''ll be passing through some pirate-infested waters soon. Should we be concerned?', '<p><strong>Captain</strong>, one more thing. We&#39;ll be passing through some pirate-infested waters soon. Should we be concerned?</p>', '0000000000000000000000000000000000000000000000000000000000002008', -1, 'approved', '2023-02-27 18:35:04.089000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200c', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Alright, engineer. Let''s schedule a time for you to do a full inspection. I want to make sure everything is shipshape before we set sail.', '<p>Alright, engineer. Let&#39;s schedule a time for you to do a full inspection. I want to make sure everything is shipshape before we set sail.</p>', '0000000000000000000000000000000000000000000000000000000000002009', 1, 'approved', '2023-02-27 18:34:35.679000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200d', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001011', 'Captain, I''ve plotted our course, and I suggest we take the eastern route. It''ll take us a bit longer, but we''ll avoid any bad weather.', '<p>Captain, I&#39;ve plotted our course, and I suggest we take the eastern route. It&#39;ll take us a bit longer, but we&#39;ll avoid any bad weather.</p>', '0000000000000000000000000000000000000000000000000000000000002003', 2, 'approved', '2023-02-27 18:28:51.050000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200e', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001003', 'I can whip up some extra spicy food to make sure any pirates who try to board us get a taste of their own medicine! 🤣', '<p>I can whip up some extra spicy food to make sure any pirates who try to board us get a taste of their own medicine! 🤣</p>', '000000000000000000000000000000000000000000000000000000000000200a', 3, 'approved', '2023-02-27 18:36:37.704000', false, null, null),
        ('000000000000000000000000000000000000000000000000000000000000200f', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001001', 'Let''s hope it doesn''t come to that, cook. But it''s good to know we have you on our side. Alright, everyone, let''s get to work. We''ve got a long journey ahead of us!', '<p>Let&#39;s hope it doesn&#39;t come to that, cook. But it&#39;s good to know we have you on our side.</p><p>Alright, everyone, let&#39;s get to work. We&#39;ve got a long journey ahead of us!</p>', '000000000000000000000000000000000000000000000000000000000000200e', 0, 'approved', '2023-02-27 18:37:24.355000', false, null, null),
        ('0000000000000000000000000000000000000000000000000000000000002010', 'localhost:8000', '/', '0000000000000000000000000000000000000000000000000000000000001003', 'We''ve got enough food 🍖 and water 🚰 to last us for the whole journey, captain. But I do have a request. Could we get some fresh vegetables 🥕🥔🍅 and fruit 🍎🍐🍌 at our next port stop? It''ll help us avoid scurvy.', '<p>We&#39;ve got enough food 🍖 and water 🚰 to last us for the whole journey, captain. But I do have a request. Could we get some fresh vegetables 🥕🥔🍅 and fruit 🍎🍐🍌 at our next port stop? It&#39;ll help us avoid scurvy.</p>', '0000000000000000000000000000000000000000000000000000000000002006', 2, 'approved', '2023-02-27 18:32:46.248000', false, null, null);

insert into votes(commenthex, commenterhex, direction, votedate)
    values
        ('0000000000000000000000000000000000000000000000000000000000002001', '0000000000000000000000000000000000000000000000000000000000001003', 1, '2023-02-27 18:38:03.542483'),
        ('0000000000000000000000000000000000000000000000000000000000002001', '0000000000000000000000000000000000000000000000000000000000001002', 1, '2023-02-27 18:38:15.842977'),
        ('0000000000000000000000000000000000000000000000000000000000002001', '0000000000000000000000000000000000000000000000000000000000001004', 1, '2023-02-27 18:38:24.880867'),
        ('0000000000000000000000000000000000000000000000000000000000002001', '0000000000000000000000000000000000000000000000000000000000001011', 1, '2023-02-27 18:39:20.157638'),
        ('0000000000000000000000000000000000000000000000000000000000002010', '0000000000000000000000000000000000000000000000000000000000001011', 1, '2023-02-27 18:39:23.238001'),
        ('000000000000000000000000000000000000000000000000000000000000200e', '0000000000000000000000000000000000000000000000000000000000001011', 1, '2023-02-27 18:39:26.023446'),
        ('000000000000000000000000000000000000000000000000000000000000200c', '0000000000000000000000000000000000000000000000000000000000001011', 1, '2023-02-27 18:39:30.894800'),
        ('000000000000000000000000000000000000000000000000000000000000200e', '0000000000000000000000000000000000000000000000000000000000001001', 1, '2023-02-27 18:39:45.074171'),
        ('0000000000000000000000000000000000000000000000000000000000002010', '0000000000000000000000000000000000000000000000000000000000001001', 1, '2023-02-27 18:40:04.176863'),
        ('000000000000000000000000000000000000000000000000000000000000200d', '0000000000000000000000000000000000000000000000000000000000001001', 1, '2023-02-27 18:40:09.569962'),
        ('000000000000000000000000000000000000000000000000000000000000200b', '0000000000000000000000000000000000000000000000000000000000001001', -1, '2023-02-27 18:40:20.406853'),
        ('000000000000000000000000000000000000000000000000000000000000200e', '0000000000000000000000000000000000000000000000000000000000001004', 1, '2023-02-27 18:40:31.890010'),
        ('000000000000000000000000000000000000000000000000000000000000200d', '0000000000000000000000000000000000000000000000000000000000001004', 1, '2023-02-27 18:40:45.479813');
