-- Make admin user root@comentario.app / admin
insert into owners(ownerhex, email, name, passwordhash, confirmedemail, joindate)
    values
        ('05878df7449326d8ad6d2fdc5c3d703fb04c72ea1a0efaa5e02ea2c3855a42e2', 'root@comentario.app', 'Admin User',
         '$2a$10$WLeCsMc7z7vSdococ9FLF.9FdcrIsJAQCeCSYFbiqFk8qRVQ/pqRK', 'true', '2023-01-17 17:55:47.008851');

insert into commenters (commenterhex, email, name, link, photo, provider, joindate, state, passwordhash)
    values
        ('d668b826923228bd75c64a8b99cc3d8dfa4179dd7e8121eaeced9eee8d4e20db', 'root@comentario.app', 'Admin User', 'undefined', 'undefined', 'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$WLeCsMc7z7vSdococ9FLF.9FdcrIsJAQCeCSYFbiqFk8qRVQ/pqRK'),
        ('296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb03', 'user@example.com', 'Test User', 'undefined', 'undefined', 'commento', '2023-01-18 16:52:04.541982', 'ok', '$2a$10$3w4LEMCh1iKwJC2uMGCP0eb0BRULg77KmnZuvnlGBMs4ALDbJ5Syy'),
        ('296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb04', 'user2@example.com', 'Another One', 'https://wikipedia.org/', 'undefined', 'commento', '2023-01-18 16:52:04.541982', 'ok', '$2a$10$3w4LEMCh1iKwJC2uMGCP0eb0BRULg77KmnZuvnlGBMs4ALDbJ5Syy');

insert into domains(domain, ownerhex, name, creationdate, state, importedcomments, autospamfilter,
                           requiremoderation, requireidentification, viewsthismonth, moderateallanonymous,
                           emailnotificationpolicy, commentoprovider, googleprovider, twitterprovider, githubprovider,
                           gitlabprovider, ssoprovider, ssosecret, ssourl, defaultsortpolicy)
    values
        ('localhost:8000', '05878df7449326d8ad6d2fdc5c3d703fb04c72ea1a0efaa5e02ea2c3855a42e2', 'Test Domain',
         '2023-01-17 17:56:10.966890', 'unfrozen', 'false', true, false, true, 0, true, 'pending-moderation', true, true,
         true, true, true, false, '', '', 'score-desc');

insert into emails (email, unsubscribesecrethex, lastemailnotificationdate, pendingemails, sendreplynotifications, sendmoderatornotifications)
    values
        ('root@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcb8', '2023-01-17 17:55:46.953534', 0, false, true),
        ('user@example.com', '2690cab8b021140dfb7d6a56ac60ac49cae3e4706a2e90b4b5645584f59451c7', '2023-01-18 16:52:04.448105', 0, false, true),
        ('user2@example.com', '2690cab8b021140dfb7d6a56ac60ac49cae3e4706a2e90b4b5645584f59451c8', '2023-01-18 16:52:04.448105', 0, false, true);

insert into moderators (domain, email, adddate)
    values
        ('localhost:8000', 'root@comentario.app', '2023-01-17 17:56:10.968427');

insert into pages (domain, path, islocked, commentcount, stickycommenthex, title)
    values
        ('localhost:8000', '/', false, 1, 'none', '');

insert into comments (commenthex, domain, path, commenterhex, markdown, html, parenthex, score, state, creationdate, deleted, deleterhex, deletiondate)
    values
        ('c3ad9084f698f6b4014b3d126f548dffdb7e908806ab630dd512895b0543b779', 'localhost:8000', '/', '296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb04', 'Hello back', '<p>Hello back', '805dca5d3ff5b7131c28c7054325b8d7aac7062145422438902911d9d50bd03b', 0, 'approved', '2023-01-18 16:52:36.161440', false, null, null),
        ('c3ad9084f698f6b4014b3d126f548dffdb7e908806ab630dd512895b0543b778', 'localhost:8000', '/', '296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb03', 'Hello **hello**! How are you? Long time no see :-P', '<p>Hello <strong>hello</strong>! How are you? Long time no see :-P</p>', '805dca5d3ff5b7131c28c7054325b8d7aac7062145422438902911d9d50bd03b', 0, 'approved', '2023-01-18 16:52:36.161440', false, null, null),
        ('3f41cdde52f24cbf171a129b57013382f959287e40a3e73b1f1433dbf7262754', 'localhost:8000', '/', 'd668b826923228bd75c64a8b99cc3d8dfa4179dd7e8121eaeced9eee8d4e20db', 'What a great website!', '<p>What a great website!</p>', 'root', 0, 'approved', '2023-01-18 16:44:55.002613', false, null, null),
        ('805dca5d3ff5b7131c28c7054325b8d7aac7062145422438902911d9d50bd03b', 'localhost:8000', '/', 'd668b826923228bd75c64a8b99cc3d8dfa4179dd7e8121eaeced9eee8d4e20db', 'Hey there!', '<p>Hey there!</p>', 'root', 2, 'approved', '2023-01-17 18:28:10.767326', false, null, null),
        ('ba845f476b2aec946f73e1b80bf43a28b258376c7e48b015d7d0332ba09b1bd3', 'localhost:8000', '/', '296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb04', 'But I must explain to you how all this mistaken idea of denouncing pleasure and praising pain was born and I will give you a complete account of the system, and expound the actual teachings of the great explorer of the truth, the master-builder of human happiness. No one rejects, dislikes, or avoids pleasure itself, *because it is pleasure, but because those who do not know how to pursue pleasure rationally encounter consequences that are extremely painful*.', '<p>But I must explain to you how all this mistaken idea of denouncing pleasure and praising pain was born and I will give you a complete account of the system, and expound the actual teachings of the great explorer of the truth, the master-builder of human happiness. No one rejects, dislikes, or avoids pleasure itself, <em>because it is pleasure, but because those who do not know how to pursue pleasure rationally encounter consequences that are extremely painful</em>.</p>', 'c3ad9084f698f6b4014b3d126f548dffdb7e908806ab630dd512895b0543b778', 0, 'approved', '2023-01-18 17:36:17.635321', false, null, null),
        ('121a85e7dcf74276aaef3f3ff7656f0b2fc2da77e35d04aa25d53ce8a4668e0f', 'localhost:8000', '/', '296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb03', 'I wholeheartedly agree, **Another**!', '<p>I wholeheartedly agree, <strong>Another</strong>!</p>', 'ba845f476b2aec946f73e1b80bf43a28b258376c7e48b015d7d0332ba09b1bd3', 0, 'approved', '2023-01-18 17:37:56.789142', false, null, null);

insert into votes (commenthex, commenterhex, direction, votedate)
    values
        ('805dca5d3ff5b7131c28c7054325b8d7aac7062145422438902911d9d50bd03b', '296c71d3d952378bcf49da722de949396b6439caf4c426274443e81093a3cb03', 1, '2023-01-18 16:52:50.830503');
