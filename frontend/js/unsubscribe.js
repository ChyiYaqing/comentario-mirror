(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    let e;

    // Update the email records.
    global.emailUpdate = function () {
        $('.err').text('');
        $('.msg').text('');
        e.sendModeratorNotifications = $('#moderator').is(':checked');
        e.sendReplyNotifications = $('#reply').is(':checked');

        const json = {
            'email': e,
        };

        global.buttonDisable('#save-button');
        global.post(global.origin + '/api/email/update', json, function (resp) {
            global.buttonEnable('#save-button');
            if (!resp.success) {
                $('.err').text(resp.message);
                return;
            }

            $('.msg').text('Successfully updated!');
        });
    }

    // Checks the unsubscribeSecretHex token to retrieve current settings.
    global.emailGet = function () {
        $('.err').text('');
        $('.msg').text('');
        const json = {
            'unsubscribeSecretHex': global.paramGet('unsubscribeSecretHex'),
        };

        global.post(global.origin + '/api/email/get', json, function (resp) {
            $('.loading').hide();
            if (!resp.success) {
                $('.err').text(resp.message);
                return;
            }

            e = resp.email;
            $('#email').text(e.email);
            $('#moderator').prop('checked', e.sendModeratorNotifications);
            $('#reply').prop('checked', e.sendReplyNotifications);
            $('.checkboxes').attr('style', '');
        });
    };

}(window.commento, document));
