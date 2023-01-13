(function (global, document) {
    'use strict';

    // Get self details.
    global.selfGet = function (callback) {
        const json = {
            'ownerToken': global.cookieGet('commentoOwnerToken'),
        };

        if (json.ownerToken === undefined) {
            document.location = global.origin + '/login';
            return;
        }

        global.post(global.origin + '/api/owner/self', json, function (resp) {
            if (!resp.success || !resp.loggedIn) {
                global.cookieDelete('commentoOwnerToken');
                document.location = global.origin + '/login';
                return;
            }

            global.owner = resp.owner;
            callback();
        });
    };

}(window.commento, document));
