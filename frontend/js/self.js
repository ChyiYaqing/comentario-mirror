(function (global, document) {
    'use strict';

    // Get self details.
    global.selfGet = function (callback) {
        const json = {ownerToken: global.cookieGet('comentarioOwnerToken')};

        if (json.ownerToken === undefined) {
            document.location = `${global.origin}/login`;
            return;
        }

        global.post(`${global.origin}/api/owner/self`, json, function (resp) {
            if (!resp.success || !resp.loggedIn) {
                global.cookieDelete('comentarioOwnerToken');
                document.location = `${global.origin}/login`;
                return;
            }

            global.owner = resp.owner;
            callback();
        });
    };

}(window.comentario, document));
