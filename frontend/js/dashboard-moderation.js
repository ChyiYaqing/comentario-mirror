(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Opens the moderatiosn settings window.
    global.moderationOpen = function () {
        $('.view').hide();
        $('#moderation-view').show();
    };

    // Adds a moderator.
    global.moderatorNewHandler = function () {
        const data = global.dashboard.$data;
        const email = $('#new-mod').val();
        const json = {
            ownerToken: global.cookieGet('comentarioOwnerToken'),
            domain: data.domains[data.cd].domain,
            email: email,
        };

        let idx = -1;
        for (let i = 0; i < data.domains[data.cd].moderators.length; i++) {
            if (data.domains[data.cd].moderators[i].email === email) {
                idx = i;
                break;
            }
        }

        if (idx === -1) {
            data.domains[data.cd].moderators.push({email: email, timeAgo: 'just now'});
            global.buttonDisable('#new-mod-button');
            global.post(`${global.origin}/api/domain/moderator/new`, json, function (resp) {
                global.buttonEnable('#new-mod-button');

                if (!resp.success) {
                    global.globalErrorShow(resp.message);
                    return;
                }

                global.globalOKShow('Added a new moderator!');
                const nm = $('#new-mod');
                nm.val('');
                nm.focus();
            });
        } else {
            global.globalErrorShow('Already a moderator.');
        }
    };

    // Deletes a moderator.
    global.moderatorDeleteHandler = function (email) {
        const data = global.dashboard.$data;

        const json = {
            ownerToken: global.cookieGet('comentarioOwnerToken'),
            domain: data.domains[data.cd].domain,
            email: email,
        };

        let idx = -1;
        for (let i = 0; i < data.domains[data.cd].moderators.length; i++) {
            if (data.domains[data.cd].moderators[i].email === email) {
                idx = i;
                break;
            }
        }

        if (idx !== -1) {
            data.domains[data.cd].moderators.splice(idx, 1);
            global.post(`${global.origin}/api/domain/moderator/delete`, json, function (resp) {
                if (!resp.success) {
                    global.globalErrorShow(resp.message);
                    return;
                }

                global.globalOKShow('Removed!');
            });
        }
    };

}(window.comentario, document));
