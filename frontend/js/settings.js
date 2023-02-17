(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    global.vueConstruct = function (callback) {
        const reactiveData = {
            hasSource: global.owner.hasSource,
            lastFour: global.owner.lastFour,
        };

        global.settings = new Vue({
            el: '#settings',
            data: reactiveData,
        });

        if (callback !== undefined) {
            callback();
        }
    };

    global.settingShow = function (setting) {
        $('.pane-setting').removeClass('selected');
        $('.view').hide();
        $(`#${setting}`).addClass('selected');
        $(`#${setting}-view`).show();
    };

    global.deleteOwnerHandler = function () {
        if (!confirm('Are you absolutely sure you want to delete your account?')) {
            return;
        }

        const json = {ownerToken: global.cookieGet('comentarioOwnerToken')};
        const delBtn = $('#delete-owner-button');
        delBtn.prop('disabled', true);
        delBtn.text('Deleting...');
        global.post(`${global.origin}/api/owner/delete`, json, function (resp) {
            if (!resp.success) {
                delBtn.prop('disabled', false);
                delBtn.text('Delete Account');
                global.globalErrorShow(resp.message);
                $('#error-message').text(resp.message);
                return;
            }

            global.cookieDelete('comentarioOwnerToken');
            document.location = `${global.origin}/login?deleted=true`;
        });
    };

}(window.comentario, document));
