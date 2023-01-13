// noinspection DuplicatedCode

(function (global, document) {
    'use strict';

    // Do not use other files like utils.js and http.js in the gulpfile to build
    // commento.js for the following reasons:
    //   - We don't use jQuery in the actual JavaScript payload because we need
    //     to be lightweight.
    //   - They pollute the global/window namespace (with global.post, etc.).
    //     That's NOT fine when we expect them to source our JavaScript. For example,
    //     the user may have their own window.post defined. We don't want to
    //     override that.

    let ID_ROOT = 'commento';

    const ID_MAIN_AREA = 'commento-main-area';
    const ID_LOGIN = 'commento-login';
    const ID_LOGIN_BOX_CONTAINER = 'commento-login-box-container';
    const ID_LOGIN_BOX = 'commento-login-box';
    const ID_LOGIN_BOX_EMAIL_SUBTITLE = 'commento-login-box-email-subtitle';
    const ID_LOGIN_BOX_EMAIL_INPUT = 'commento-login-box-email-input';
    const ID_LOGIN_BOX_PASSWORD_INPUT = 'commento-login-box-password-input';
    const ID_LOGIN_BOX_NAME_INPUT = 'commento-login-box-name-input';
    const ID_LOGIN_BOX_WEBSITE_INPUT = 'commento-login-box-website-input';
    const ID_LOGIN_BOX_EMAIL_BUTTON = 'commento-login-box-email-button';
    const ID_LOGIN_BOX_FORGOT_LINK_CONTAINER = 'commento-login-box-forgot-link-container';
    const ID_LOGIN_BOX_LOGIN_LINK_CONTAINER = 'commento-login-box-login-link-container';
    const ID_LOGIN_BOX_SSO_PRETEXT = 'commento-login-box-sso-pretext';
    const ID_LOGIN_BOX_SSO_BUTTON_CONTAINER = 'commento-login-box-sso-button-container';
    const ID_LOGIN_BOX_HR1 = 'commento-login-box-hr1';
    const ID_LOGIN_BOX_OAUTH_PRETEXT = 'commento-login-box-oauth-pretext';
    const ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER = 'commento-login-box-oauth-buttons-container';
    const ID_LOGIN_BOX_HR2 = 'commento-login-box-hr2';
    const ID_MOD_TOOLS = 'commento-mod-tools';
    const ID_MOD_TOOLS_LOCK_BUTTON = 'commento-mod-tools-lock-button';
    const ID_ERROR = 'commento-error';
    const ID_LOGGED_CONTAINER = 'commento-logged-container';
    const ID_PRE_COMMENTS_AREA = 'commento-pre-comments-area';
    const ID_COMMENTS_AREA = 'commento-comments-area';
    const ID_SUPER_CONTAINER = 'commento-textarea-super-container-';
    const ID_TEXTAREA_CONTAINER = 'commento-textarea-container-';
    const ID_TEXTAREA = 'commento-textarea-';
    const ID_ANONYMOUS_CHECKBOX = 'commento-anonymous-checkbox-';
    const ID_SORT_POLICY = 'commento-sort-policy-';
    const ID_CARD = 'commento-comment-card-';
    const ID_BODY = 'commento-comment-body-';
    const ID_TEXT = 'commento-comment-text-';
    const ID_SUBTITLE = 'commento-comment-subtitle-';
    const ID_TIMEAGO = 'commento-comment-timeago-';
    const ID_SCORE = 'commento-comment-score-';
    const ID_OPTIONS = 'commento-comment-options-';
    const ID_EDIT = 'commento-comment-edit-';
    const ID_REPLY = 'commento-comment-reply-';
    const ID_COLLAPSE = 'commento-comment-collapse-';
    const ID_UPVOTE = 'commento-comment-upvote-';
    const ID_DOWNVOTE = 'commento-comment-downvote-';
    const ID_APPROVE = 'commento-comment-approve-';
    const ID_REMOVE = 'commento-comment-remove-';
    const ID_STICKY = 'commento-comment-sticky-';
    const ID_CHILDREN = 'commento-comment-children-';
    const ID_CONTENTS = 'commento-comment-contents-';
    const ID_NAME = 'commento-comment-name-';
    const ID_SUBMIT_BUTTON = 'commento-submit-button-';
    const ID_MARKDOWN_BUTTON = 'commento-markdown-button-';
    const ID_MARKDOWN_HELP = 'commento-markdown-help-';
    const ID_FOOTER = 'commento-footer';

    const origin = '[[[.Origin]]]';
    const cdn = '[[[.CdnPrefix]]]';

    let root = null;
    let pageId = parent.location.pathname;
    let cssOverride;
    let noFonts;
    let hideDeleted;
    let autoInit;
    let isAuthenticated = false;
    let comments = [];
    let commentsMap = {};
    let commenters = {};
    let requireIdentification = true;
    let isModerator = false;
    let isFrozen = false;
    let chosenAnonymous = false;
    let isLocked = false;
    let stickyCommentHex = 'none';
    let shownReply = {};
    let shownEdit = {};
    let configuredOauths = {};
    let anonymousOnly = false;
    let popupBoxType = 'login';
    let oauthButtonsShown = false;
    let sortPolicy = 'score-desc';
    let selfHex = undefined;
    let mobileView = null;

    function $(id) {
        return document.getElementById(id);
    }

    function tags(tag) {
        return document.getElementsByTagName(tag);
    }

    function prepend(root, el) {
        root.prepend(el);
    }

    function append(root, el) {
        root.appendChild(el);
    }

    function insertAfter(el1, el2) {
        el1.parentNode.insertBefore(el2, el1.nextSibling);
    }

    function classAdd(el, cls) {
        el.classList.add(`commento-${cls}`);
    }

    function classRemove(el, cls) {
        if (el !== null) {
            el.classList.remove(`commento-${cls}`);
        }
    }

    function create(el, id) {
        const e = document.createElement(el);
        if (id) {
            e.id = id;
        }
        return e;
    }

    function remove(el) {
        if (el !== null) {
            el.parentNode.removeChild(el);
        }
    }

    function getAttr(node, a) {
        const attr = node.attributes[a];
        return attr === void 0 ? undefined : attr.value;
    }

    function removeAllEventListeners(node) {
        if (node !== null) {
            const replacement = node.cloneNode(true);
            if (node.parentNode !== null) {
                node.parentNode.replaceChild(replacement, node);
                return replacement;
            }
        }
        return node;
    }

    function onclick(node, f, arg) {
        node.addEventListener('click', function () {
            f(arg);
        }, false);
    }

    function onload(node, f, arg) {
        node.addEventListener('load', function () {
            f(arg);
        });
    }

    function setAttr(node, values) {
        Object.keys(values).forEach(k => node.setAttribute(k, values[k]));
    }

    function post(url, data, callback) {
        const xmlDoc = new XMLHttpRequest();

        xmlDoc.open('POST', url, true);
        xmlDoc.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        xmlDoc.onload = function () {
            callback(JSON.parse(xmlDoc.response));
        };

        xmlDoc.send(JSON.stringify(data));
    }

    function get(url, callback) {
        const xmlDoc = new XMLHttpRequest();

        xmlDoc.open('GET', url, true);
        xmlDoc.onload = function () {
            callback(JSON.parse(xmlDoc.response));
        };

        xmlDoc.send(null);
    }

    function call(callback) {
        if (typeof (callback) === 'function') {
            callback();
        }
    }

    function cookieGet(name) {
        const c = `; ${document.cookie}`;
        const x = c.split(`; ${name}=`);
        if (x.length === 2) {
            return x.pop().split(';').shift();
        }
    }

    function cookieSet(name, value) {
        const date = new Date();
        date.setTime(date.getTime() + (365 * 24 * 60 * 60 * 1000));
        document.cookie = `${name}=${value}; expires=${date.toUTCString()}; path=/`;
    }

    function commenterTokenGet() {
        const commenterToken = cookieGet('commentoCommenterToken');
        return commenterToken === undefined ? 'anonymous' : commenterToken;
    }

    global.logout = function () {
        cookieSet('commentoCommenterToken', 'anonymous');
        isAuthenticated = false;
        isModerator = false;
        selfHex = undefined;
        refreshAll();
    }

    function profileEdit() {
        window.open(`${origin}/profile?commenterToken=${commenterTokenGet()}`, '_blank');
    }

    function notificationSettings(unsubscribeSecretHex) {
        window.open(`${origin}/unsubscribe?unsubscribeSecretHex=${unsubscribeSecretHex}`, '_blank');
    }

    function selfLoad(commenter, email) {
        commenters[commenter.commenterHex] = commenter;
        selfHex = commenter.commenterHex;

        const loggedContainer = create('div', ID_LOGGED_CONTAINER);
        const loggedInAs = create('div');
        const name = commenter.link !== 'undefined' ? create('a') : create('div');
        let avatar;
        const notificationSettingsButton = create('div');
        const profileEditButton = create('div');
        const logoutButton = create('div');
        const color = colorGet(`${commenter.commenterHex}-${commenter.name}`);

        classAdd(loggedContainer, 'logged-container');
        classAdd(loggedInAs, 'logged-in-as');
        classAdd(name, 'name');
        classAdd(notificationSettingsButton, 'profile-button');
        classAdd(profileEditButton, 'profile-button');
        classAdd(logoutButton, 'profile-button');

        name.innerText = commenter.name;
        notificationSettingsButton.innerText = 'Notification Settings';
        profileEditButton.innerText = 'Edit Profile';
        logoutButton.innerText = 'Logout';

        onclick(logoutButton, global.logout);
        onclick(notificationSettingsButton, notificationSettings, email.unsubscribeSecretHex);
        onclick(profileEditButton, profileEdit);

        setAttr(loggedContainer, {style: 'display: none'});
        if (commenter.link !== 'undefined') {
            setAttr(name, {href: commenter.link});
        }
        if (commenter.photo === 'undefined') {
            avatar = create('div');
            avatar.style['background'] = color;
            avatar.innerHTML = commenter.name[0].toUpperCase();
            classAdd(avatar, 'avatar');
        } else {
            avatar = create('img');
            setAttr(
                avatar,
                {src: `${cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`, loading: 'lazy', alt: ''});
            classAdd(avatar, 'avatar-img');
        }

        append(loggedInAs, avatar);
        append(loggedInAs, name);
        append(loggedContainer, loggedInAs);
        append(loggedContainer, logoutButton);
        if (commenter.provider === 'commento') {
            append(loggedContainer, profileEditButton);
        }
        append(loggedContainer, notificationSettingsButton);
        prepend(root, loggedContainer);

        isAuthenticated = true;
    }

    function selfGet(callback) {
        const commenterToken = commenterTokenGet();
        if (commenterToken === 'anonymous') {
            isAuthenticated = false;
            call(callback);
            return;
        }

        const json = {commenterToken: commenterTokenGet()};

        post(`${origin}/api/commenter/self`, json, resp => {
            if (!resp.success) {
                cookieSet('commentoCommenterToken', 'anonymous');
                call(callback);
                return;
            }

            selfLoad(resp.commenter, resp.email);
            allShow();

            call(callback);
        });
    }

    function cssLoad(file, f) {
        const link = create('link');
        const head = document.getElementsByTagName('head')[0];
        setAttr(link, {href: file, rel: 'stylesheet', type: 'text/css'});
        onload(link, f);
        append(head, link);
    }

    function footerLoad() {
        const footer = create('div', ID_FOOTER);
        const aContainer = create('div');
        const a = create('a');
        const text = create('span');

        classAdd(footer, 'footer');
        classAdd(aContainer, 'logo-container');
        classAdd(a, 'logo');
        classAdd(text, 'logo-text');

        setAttr(a, {href: 'https://gitlab.com/comentario/comentario', target: '_blank'});

        text.innerText = 'Comentario 🗨';

        append(a, text);
        append(aContainer, a);
        append(footer, aContainer);

        return footer;
    }

    function commentsGet(callback) {
        const json = {
            commenterToken: commenterTokenGet(),
            domain: parent.location.host,
            path: pageId,
        };

        post(`${origin}/api/comment/list`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return;
            }

            errorHide();

            requireIdentification = resp.requireIdentification;
            isModerator = resp.isModerator;
            isFrozen = resp.isFrozen;

            isLocked = resp.attributes.isLocked;
            stickyCommentHex = resp.attributes.stickyCommentHex;

            comments = resp.comments;
            commenters = Object.assign({}, commenters, resp.commenters)
            configuredOauths = resp.configuredOauths;

            sortPolicy = resp.defaultSortPolicy;

            call(callback);
        });
    }

    function errorShow(text) {
        const el = $(ID_ERROR);
        el.innerText = text;
        setAttr(el, {style: 'display: block;'});
    }

    function errorHide() {
        setAttr($(ID_ERROR), {style: 'display: none;'});
    }

    function errorElementCreate() {
        const el = create('div', ID_ERROR);
        classAdd(el, 'error-box');
        setAttr(el, {style: 'display: none;'});
        append(root, el);
    }

    function autoExpander(el) {
        return () => {
            el.style.height = '';
            el.style.height = `${Math.min(Math.max(el.scrollHeight, 75), 400)}px`;
        }
    }

    function markdownHelpShow(id) {
        const textareaSuperContainer = $(ID_SUPER_CONTAINER + id);
        let markdownButton = $(ID_MARKDOWN_BUTTON + id);
        const markdownHelp = create('table', ID_MARKDOWN_HELP + id);
        const italicsContainer = create('tr');
        const italicsLeft = create('td');
        const italicsRight = create('td');
        const boldContainer = create('tr');
        const boldLeft = create('td');
        const boldRight = create('td');
        const codeContainer = create('tr');
        const codeLeft = create('td');
        const codeRight = create('td');
        const strikethroughContainer = create('tr');
        const strikethroughLeft = create('td');
        const strikethroughRight = create('td');
        const hyperlinkContainer = create('tr');
        const hyperlinkLeft = create('td');
        const hyperlinkRight = create('td');
        const quoteContainer = create('tr');
        const quoteLeft = create('td');
        const quoteRight = create('td');

        classAdd(markdownHelp, 'markdown-help');

        boldLeft.innerHTML = '<b>bold</b>';
        boldRight.innerHTML = 'surround text with <pre>**two asterisks**</pre>';
        italicsLeft.innerHTML = '<i>italics</i>';
        italicsRight.innerHTML = 'surround text with <pre>*asterisks*</pre>';
        codeLeft.innerHTML = '<pre>code</pre>';
        codeRight.innerHTML = 'surround text with <pre>`backticks`</pre>';
        strikethroughLeft.innerHTML = '<del>strikethrough</del>';
        strikethroughRight.innerHTML = 'surround text with <pre>~~two tilde characters~~</pre>';
        hyperlinkLeft.innerHTML = '<a href="https://example.com">hyperlink</a>';
        hyperlinkRight.innerHTML = '<pre>[hyperlink](https://example.com)</pre> or just a bare URL';
        quoteLeft.innerHTML = '<blockquote>quote</blockquote>';
        quoteRight.innerHTML = 'prefix with <pre>&gt;</pre>';

        markdownButton = removeAllEventListeners(markdownButton);
        onclick(markdownButton, markdownHelpHide, id);

        append(italicsContainer, italicsLeft);
        append(italicsContainer, italicsRight);
        append(markdownHelp, italicsContainer);
        append(boldContainer, boldLeft);
        append(boldContainer, boldRight);
        append(markdownHelp, boldContainer);
        append(hyperlinkContainer, hyperlinkLeft);
        append(hyperlinkContainer, hyperlinkRight);
        append(markdownHelp, hyperlinkContainer);
        append(codeContainer, codeLeft);
        append(codeContainer, codeRight);
        append(markdownHelp, codeContainer);
        append(strikethroughContainer, strikethroughLeft);
        append(strikethroughContainer, strikethroughRight);
        append(markdownHelp, strikethroughContainer);
        append(quoteContainer, quoteLeft);
        append(quoteContainer, quoteRight);
        append(markdownHelp, quoteContainer);
        append(textareaSuperContainer, markdownHelp);
    }

    function markdownHelpHide(id) {
        let markdownButton = $(ID_MARKDOWN_BUTTON + id);
        const markdownHelp = $(ID_MARKDOWN_HELP + id);

        markdownButton = removeAllEventListeners(markdownButton);
        onclick(markdownButton, markdownHelpShow, id);

        remove(markdownHelp);
    }

    function textareaCreate(id, edit) {
        const textareaSuperContainer     = create('div',      ID_SUPER_CONTAINER    + id);
        const textareaContainer          = create('div',      ID_TEXTAREA_CONTAINER + id);
        const textarea                   = create('textarea', ID_TEXTAREA           + id);
        const anonymousCheckboxContainer = create('div');
        const anonymousCheckbox          = create('input',    ID_ANONYMOUS_CHECKBOX + id);
        const anonymousCheckboxLabel     = create('label');
        const submitButton               = create('button',   ID_SUBMIT_BUTTON      + id);
        const markdownButton             = create('a',        ID_MARKDOWN_BUTTON    + id);

        classAdd(textareaContainer, 'textarea-container');
        classAdd(anonymousCheckboxContainer, 'round-check');
        classAdd(anonymousCheckboxContainer, 'anonymous-checkbox-container');
        classAdd(submitButton, 'button');
        classAdd(submitButton, 'submit-button');
        classAdd(markdownButton, 'markdown-button');
        classAdd(textareaSuperContainer, 'button-margin');

        setAttr(textarea, {placeholder: 'Add a comment'});
        setAttr(anonymousCheckbox, {type: 'checkbox'});
        setAttr(anonymousCheckboxLabel, {for: ID_ANONYMOUS_CHECKBOX + id});

        anonymousCheckboxLabel.innerText = 'Comment anonymously';
        submitButton.innerText = edit ? 'Save Changes' : 'Add Comment';
        markdownButton.innerHTML = '<b>M &#8595;</b> &nbsp; Markdown';

        if (anonymousOnly) {
            anonymousCheckbox.checked = true;
            anonymousCheckbox.setAttribute('disabled', true);
        }

        textarea.oninput = autoExpander(textarea);
        onclick(submitButton, edit ? commentEdit : submitAccountDecide, id);
        onclick(markdownButton, markdownHelpShow, id);

        append(textareaContainer, textarea);
        append(textareaSuperContainer, textareaContainer);
        append(anonymousCheckboxContainer, anonymousCheckbox);
        append(anonymousCheckboxContainer, anonymousCheckboxLabel);
        append(textareaSuperContainer, submitButton);
        if (!requireIdentification && edit !== true) {
            append(textareaSuperContainer, anonymousCheckboxContainer);
        }
        append(textareaSuperContainer, markdownButton);

        return textareaSuperContainer;
    }

    const sortPolicyNames = {
        'score-desc': 'Upvotes',
        'creationdate-desc': 'Newest',
        'creationdate-asc': 'Oldest',
    };

    function sortPolicyApply(policy) {
        classRemove($(ID_SORT_POLICY + sortPolicy), 'sort-policy-button-selected');

        const commentsArea = $(ID_COMMENTS_AREA);
        commentsArea.innerHTML = '';
        sortPolicy = policy;
        const cards = commentsRecurse(parentMap(comments), 'root');
        if (cards) {
            append(commentsArea, cards);
        }

        classAdd($(ID_SORT_POLICY + policy), 'sort-policy-button-selected');
    }

    function sortPolicyBox() {
        const sortPolicyButtonsContainer = create('div');
        const sortPolicyButtons = create('div');

        classAdd(sortPolicyButtonsContainer, 'sort-policy-buttons-container');
        classAdd(sortPolicyButtons, 'sort-policy-buttons');

        for (let sp in sortPolicyNames) {
            const sortPolicyButton = create('a', ID_SORT_POLICY + sp);
            classAdd(sortPolicyButton, 'sort-policy-button');
            if (sp === sortPolicy) {
                classAdd(sortPolicyButton, 'sort-policy-button-selected');
            }
            sortPolicyButton.innerText = sortPolicyNames[sp];
            onclick(sortPolicyButton, sortPolicyApply, sp);
            append(sortPolicyButtons, sortPolicyButton)
        }

        append(sortPolicyButtonsContainer, sortPolicyButtons);

        return sortPolicyButtonsContainer
    }

    function rootCreate(callback) {
        const login           = create('div', ID_LOGIN);
        const loginText       = create('div');
        const mainArea        = $(ID_MAIN_AREA);
        const preCommentsArea = create('div', ID_PRE_COMMENTS_AREA);
        const commentsArea    = create('div', ID_COMMENTS_AREA);

        classAdd(login, 'login');
        classAdd(loginText, 'login-text');
        classAdd(commentsArea, 'comments');

        loginText.innerText = 'Login';
        commentsArea.innerHTML = '';

        onclick(loginText, global.loginBoxShow, null);

        let numOauthConfigured = 0;
        Object.keys(configuredOauths).forEach(key => configuredOauths[key] && numOauthConfigured++);
        if (numOauthConfigured > 0) {
            append(login, loginText);
        } else if (!requireIdentification) {
            anonymousOnly = true;
        }

        if (isLocked || isFrozen) {
            if (isAuthenticated || chosenAnonymous) {
                append(mainArea, messageCreate('This thread is locked. You cannot add new comments.'));
                remove($(ID_LOGIN));
            } else {
                append(mainArea, login);
                append(mainArea, textareaCreate('root'));
            }
        } else {
            if (!isAuthenticated) {
                append(mainArea, login);
            } else {
                remove($(ID_LOGIN));
            }
            append(mainArea, textareaCreate('root'));
        }

        if (comments.length > 0) {
            append(mainArea, sortPolicyBox());
        }
        append(mainArea, preCommentsArea);
        append(mainArea, commentsArea);
        append(root, mainArea);
        call(callback);
    }

    function messageCreate(text) {
        const msg = create('div');
        classAdd(msg, 'moderation-notice');
        msg.innerText = text;
        return msg;
    }

    global.commentNew = (id, commenterToken, appendCard, callback) => {
        const textareaSuperContainer = $(ID_SUPER_CONTAINER + id);
        const textarea = $(ID_TEXTAREA + id);
        const replyButton = $(ID_REPLY + id);

        const markdown = textarea.value;

        if (markdown === '') {
            classAdd(textarea, 'red-border');
            return;
        }

        classRemove(textarea, 'red-border');

        const json = {
            commenterToken: commenterToken,
            domain: parent.location.host,
            path: pageId,
            parentHex: id,
            markdown: markdown,
        };

        post(`${origin}/api/comment/new`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return;
            }

            errorHide();

            let message = '';
            switch (resp.state) {
            case 'unapproved':
                message = 'Your comment is under moderation.';
                break;
            case 'flagged':
                message = 'Your comment was flagged as spam and is under moderation.';
                break;
            }

            if (message !== '') {
                prepend($(ID_SUPER_CONTAINER + id), messageCreate(message));
            }

            const comment = {
                commentHex: resp.commentHex,
                commenterHex: selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : selfHex,
                markdown: markdown,
                html: resp.html,
                parentHex: 'root',
                score: 0,
                state: 'approved',
                direction: 0,
                creationDate: new Date(),
            };

            const newCard = commentsRecurse({root: [comment]}, 'root');

            commentsMap[resp.commentHex] = comment;
            if (appendCard) {
                if (id !== 'root') {
                    textareaSuperContainer.replaceWith(newCard);

                    shownReply[id] = false;

                    classAdd(replyButton, 'option-reply');
                    classRemove(replyButton, 'option-cancel');

                    replyButton.title = 'Reply to this comment';

                    onclick(replyButton, global.replyShow, id)
                } else {
                    textarea.value = '';
                    insertAfter($(ID_PRE_COMMENTS_AREA), newCard);
                }
            } else if (id === 'root') {
                textarea.value = '';
            }

            call(callback);
        });
    }

    function colorGet(name) {
        const colors = [
            '#396ab1',
            '#da7c30',
            '#3e9651',
            '#cc2529',
            '#922428',
            '#6b4c9a',
            '#535154',
        ];

        let total = 0;
        for (let i = 0; i < name.length; i++) {
            total += name.charCodeAt(i);
        }
        return colors[total % colors.length];
    }

    function timeDifference(current, previous) { // thanks stackoverflow
        // Times are defined in milliseconds
        const msPerSecond = 1000;
        const msPerMinute = 60 * msPerSecond;
        const msPerHour = 60 * msPerMinute;
        const msPerDay = 24 * msPerHour;
        const msPerMonth = 30 * msPerDay;
        const msPerYear = 12 * msPerMonth;

        // Time ago thresholds
        const msJustNow = 5 * msPerSecond; // Up until 5 s
        const msMinutesAgo = 2 * msPerMinute; // Up until 2 minutes
        const msHoursAgo = 2 * msPerHour; // Up until 2 hours
        const msDaysAgo = 2 * msPerDay; // Up until 2 days
        const msMonthsAgo = 2 * msPerMonth; // Up until 2 months
        const msYearsAgo = 2 * msPerYear; // Up until 2 years

        const elapsed = current - previous;

        if (elapsed < msJustNow) {
            return 'just now';
        } else if (elapsed < msMinutesAgo) {
            return `${Math.round(elapsed / msPerSecond)} seconds ago`;
        } else if (elapsed < msHoursAgo) {
            return `${Math.round(elapsed / msPerMinute)} minutes ago`;
        } else if (elapsed < msDaysAgo) {
            return `${Math.round(elapsed / msPerHour)} hours ago`;
        } else if (elapsed < msMonthsAgo) {
            return `${Math.round(elapsed / msPerDay)} days ago`;
        } else if (elapsed < msYearsAgo) {
            return `${Math.round(elapsed / msPerMonth)} months ago`;
        } else {
            return `${Math.round(elapsed / msPerYear)} years ago`;
        }
    }

    function scorify(score) {
        return score === 1 ? `${score} point` : `${score} points`;
    }

    const sortPolicyFunctions = {
        'score-desc':        (a, b) => b.score - a.score,
        'creationdate-desc': (a, b) => a.creationDate < b.creationDate ? 1 : -1,
        'creationdate-asc':  (a, b) => a.creationDate < b.creationDate ? -1 : 1,
    };

    function commentsRecurse(parentMap, parentHex) {
        const cur = parentMap[parentHex];
        if (!cur || !cur.length) {
            return null;
        }

        cur.sort((a, b) => {
            return !a.deleted && a.commentHex === stickyCommentHex ?
                -Infinity :
                !b.deleted && b.commentHex === stickyCommentHex ?
                    Infinity :
                    sortPolicyFunctions[sortPolicy](a, b);
        });

        const curTime = (new Date()).getTime();
        const cards = create('div');
        cur.forEach(comment => {
            const commenter = commenters[comment.commenterHex];
            let avatar;
            const hex = comment.commentHex;
            const header   = create('div');
            const card     = create('div',    ID_CARD     + hex);
            const subtitle = create('div',    ID_SUBTITLE + hex);
            const timeago  = create('div',    ID_TIMEAGO  + hex);
            const score    = create('div',    ID_SCORE    + hex);
            const body     = create('div',    ID_BODY     + hex);
            const text     = create('div',    ID_TEXT     + hex);
            const options  = create('div',    ID_OPTIONS  + hex);
            const edit     = create('button', ID_EDIT     + hex);
            const reply    = create('button', ID_REPLY    + hex);
            const collapse = create('button', ID_COLLAPSE + hex);
            let   upvote   = create('button', ID_UPVOTE   + hex);
            let   downvote = create('button', ID_DOWNVOTE + hex);
            const approve  = create('button', ID_APPROVE  + hex);
            const remove   = create('button', ID_REMOVE   + hex);
            const sticky   = create('button', ID_STICKY   + hex);
            const contents = create('div',    ID_CONTENTS + hex);

            const children = commentsRecurse(parentMap, hex);
            const color = colorGet(`${comment.commenterHex}-${commenter.name}`);
            const name = create(
                commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '' ? 'a' : 'div',
                ID_NAME + hex);

            if (children) {
                children.id = ID_CHILDREN + hex;
            }

            collapse.title = 'Collapse children';
            upvote.title = 'Upvote';
            downvote.title = 'Downvote';
            edit.title = 'Edit';
            reply.title = 'Reply';
            approve.title = 'Approve';
            remove.title = 'Remove';
            if (stickyCommentHex === hex) {
                if (isModerator) {
                    sticky.title = 'Unsticky';
                } else {
                    sticky.title = 'This comment has been stickied';
                }
            } else {
                sticky.title = 'Sticky';
            }
            timeago.title = comment.creationDate.toString();

            card.style['borderLeft'] = `2px solid ${color}`;
            if (comment.deleted) {
                name.innerText = '[deleted]';
            } else {
                name.innerText = commenter.name;
            }
            text.innerHTML = comment.html;
            timeago.innerHTML = timeDifference(curTime, comment.creationDate);
            score.innerText = scorify(comment.score);

            if (commenter.photo === 'undefined') {
                avatar = create('div');
                avatar.style['background'] = color;

                if (comment.commenterHex === 'anonymous') {
                    avatar.innerHTML = '?';
                    avatar.style['font-weight'] = 'bold';
                } else {
                    avatar.innerHTML = commenter.name[0].toUpperCase();
                }

                classAdd(avatar, 'avatar');
            } else {
                avatar = create('img');
                setAttr(avatar, {src: `${cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`});
                classAdd(avatar, 'avatar-img');
            }

            classAdd(card, 'card');
            if (isModerator && comment.state !== 'approved') {
                classAdd(card, 'dark-card');
            }
            if (commenter.isModerator) {
                classAdd(name, 'moderator');
            }
            if (comment.state === 'flagged') {
                classAdd(name, 'flagged');
            }
            classAdd(header, 'header');
            classAdd(name, 'name');
            classAdd(subtitle, 'subtitle');
            classAdd(timeago, 'timeago');
            classAdd(score, 'score');
            classAdd(body, 'body');
            classAdd(options, 'options');
            if (mobileView) {
                classAdd(options, 'options-mobile');
            }
            classAdd(edit, 'option-button');
            classAdd(edit, 'option-edit');
            classAdd(reply, 'option-button');
            classAdd(reply, 'option-reply');
            classAdd(collapse, 'option-button');
            classAdd(collapse, 'option-collapse');
            classAdd(upvote, 'option-button');
            classAdd(upvote, 'option-upvote');
            classAdd(downvote, 'option-button');
            classAdd(downvote, 'option-downvote');
            classAdd(approve, 'option-button');
            classAdd(approve, 'option-approve');
            classAdd(remove, 'option-button');
            classAdd(remove, 'option-remove');
            classAdd(sticky, 'option-button');
            classAdd(sticky, stickyCommentHex === hex ? 'option-unsticky' : 'option-sticky');

            if (isAuthenticated) {
                if (comment.direction > 0) {
                    classAdd(upvote, 'upvoted');
                } else if (comment.direction < 0) {
                    classAdd(downvote, 'downvoted');
                }
            }

            onclick(edit, global.editShow, hex);
            onclick(collapse, global.commentCollapse, hex);
            onclick(approve, global.commentApprove, hex);
            onclick(remove, global.commentDelete, hex);
            onclick(sticky, global.commentSticky, hex);

            if (isAuthenticated) {
                const upDown = upDownOnclickSet(upvote, downvote, hex, comment.direction);
                upvote = upDown[0];
                downvote = upDown[1];
            } else {
                onclick(upvote, global.loginBoxShow, null);
                onclick(downvote, global.loginBoxShow, null);
            }

            onclick(reply, global.replyShow, hex);

            if (commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '') {
                setAttr(name, {href: commenter.link});
            }

            append(options, collapse);

            if (!comment.deleted) {
                append(options, downvote);
                append(options, upvote);
            }

            if (comment.commenterHex === selfHex) {
                append(options, edit);
            } else if (!comment.deleted) {
                append(options, reply);
            }

            if (!comment.deleted && (isModerator && parentHex === 'root')) {
                append(options, sticky);
            }

            if (!comment.deleted && (isModerator || comment.commenterHex === selfHex)) {
                append(options, remove);
            }

            if (isModerator && comment.state !== 'approved') {
                append(options, approve);
            }

            if (!comment.deleted && (!isModerator && stickyCommentHex === hex)) {
                append(options, sticky);
            }

            setAttr(options, {style: `width: ${(options.childNodes.length + 1) * 32}px;`});
            for (let i = 0; i < options.childNodes.length; i++) {
                setAttr(options.childNodes[i], {style: `right: ${i * 32}px;`});
            }

            append(subtitle, score);
            append(subtitle, timeago);

            if (!mobileView) {
                append(header, options);
            }
            append(header, avatar);
            append(header, name);
            append(header, subtitle);
            append(body, text);
            append(contents, body);
            if (mobileView) {
                append(contents, options);
                const optionsClearfix = create('div');
                classAdd(optionsClearfix, 'options-clearfix');
                append(contents, optionsClearfix);
            }

            if (children) {
                classAdd(children, 'body');
                append(contents, children);
            }

            append(card, header);
            append(card, contents);

            if (comment.deleted && (hideDeleted === 'true' || children === null)) {
                return;
            }

            append(cards, card);
        });

        return cards.childNodes.length === 0 ? null : cards;
    }

    global.commentApprove = commentHex => {
        const json = {
            commenterToken: commenterTokenGet(),
            commentHex: commentHex,
        };

        post(`${origin}/api/comment/approve`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return
            } else {
                errorHide();
            }

            const card = $(ID_CARD + commentHex);
            const name = $(ID_NAME + commentHex);
            const tick = $(ID_APPROVE + commentHex);

            classRemove(card, 'dark-card');
            classRemove(name, 'flagged');
            remove(tick);
        });
    }

    global.commentDelete = commentHex => {
        if (!confirm('Are you sure you want to delete this comment?')) {
            return;
        }

        const json = {
            commenterToken: commenterTokenGet(),
            commentHex: commentHex,
        };

        post(`${origin}/api/comment/delete`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return
            }

            errorHide();
            const text = $(ID_TEXT + commentHex);
            text.innerText = '[deleted]';
        });
    }

    function nameWidthFix() {
        const els = document.getElementsByClassName('commento-name');

        for (let i = 0; i < els.length; i++) {
            setAttr(els[i], {style: `max-width: ${els[i].getBoundingClientRect()['width'] + 20}px;`})
        }
    }

    function upDownOnclickSet(upvote, downvote, commentHex, direction) {
        upvote = removeAllEventListeners(upvote);
        downvote = removeAllEventListeners(downvote);

        if (direction > 0) {
            onclick(upvote, global.vote, [commentHex, [1, 0]]);
            onclick(downvote, global.vote, [commentHex, [1, -1]]);
        } else if (direction < 0) {
            onclick(upvote, global.vote, [commentHex, [-1, 1]]);
            onclick(downvote, global.vote, [commentHex, [-1, 0]]);
        } else {
            onclick(upvote, global.vote, [commentHex, [0, 1]]);
            onclick(downvote, global.vote, [commentHex, [0, -1]]);
        }

        return [upvote, downvote];
    }

    global.vote = data => {
        const commentHex = data[0];
        const oldDirection = data[1][0];
        const newDirection = data[1][1];

        let upvote = $(ID_UPVOTE + commentHex);
        let downvote = $(ID_DOWNVOTE + commentHex);
        const score = $(ID_SCORE + commentHex);

        const json = {
            commenterToken: commenterTokenGet(),
            commentHex: commentHex,
            direction: newDirection,
        };

        const upDown = upDownOnclickSet(upvote, downvote, commentHex, newDirection);
        upvote = upDown[0];
        downvote = upDown[1];

        classRemove(upvote, 'upvoted');
        classRemove(downvote, 'downvoted');
        if (newDirection > 0) {
            classAdd(upvote, 'upvoted');
        } else if (newDirection < 0) {
            classAdd(downvote, 'downvoted');
        }

        score.innerText = scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) + newDirection - oldDirection);

        post(`${origin}/api/comment/vote`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                classRemove(upvote, 'upvoted');
                classRemove(downvote, 'downvoted');
                score.innerText = scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) - newDirection + oldDirection);
                upDownOnclickSet(upvote, downvote, commentHex, oldDirection);
            } else {
                errorHide();
            }
        });
    }

    function commentEdit(id) {
        const textarea = $(ID_TEXTAREA + id);
        const markdown = textarea.value;
        if (markdown === '') {
            classAdd(textarea, 'red-border');
            return;
        }

        classRemove(textarea, 'red-border');

        const json = {
            commenterToken: commenterTokenGet(),
            commentHex: id,
            markdown: markdown,
        };

        post(`${origin}/api/comment/edit`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return;
            }

            errorHide();

            commentsMap[id].markdown = markdown;
            commentsMap[id].html = resp.html;

            let editButton = $(ID_EDIT + id);
            const textarea = $(ID_SUPER_CONTAINER + id);

            textarea.innerHTML = commentsMap[id].html;
            textarea.id = ID_TEXT + id;
            delete shownEdit[id];

            classAdd(editButton, 'option-edit');
            classRemove(editButton, 'option-cancel');

            editButton.title = 'Edit comment';

            editButton = removeAllEventListeners(editButton);
            onclick(editButton, global.editShow, id)

            let message = '';
            switch (resp.state) {
            case 'unapproved':
                message = 'Your comment is under moderation.';
                break;
            case 'flagged':
                message = 'Your comment was flagged as spam and is under moderation.';
                break;
            }

            if (message !== '') {
                prepend($(ID_SUPER_CONTAINER + id), messageCreate(message));
            }
        });
    }

    global.editShow = id => {
        if (id in shownEdit && shownEdit[id]) {
            return;
        }

        const text = $(ID_TEXT + id);
        shownEdit[id] = true;
        text.replaceWith(textareaCreate(id, true));

        const textarea = $(ID_TEXTAREA + id);
        textarea.value = commentsMap[id].markdown;

        let editButton = $(ID_EDIT + id);

        classRemove(editButton, 'option-edit');
        classAdd(editButton, 'option-cancel');

        editButton.title = 'Cancel edit';

        editButton = removeAllEventListeners(editButton);
        onclick(editButton, global.editCollapse, id);
    };

    global.editCollapse = id => {
        let editButton = $(ID_EDIT + id);
        const textarea = $(ID_SUPER_CONTAINER + id);

        textarea.innerHTML = commentsMap[id].html;
        textarea.id = ID_TEXT + id;
        delete shownEdit[id];

        classAdd(editButton, 'option-edit');
        classRemove(editButton, 'option-cancel');

        editButton.title = 'Edit comment';

        editButton = removeAllEventListeners(editButton);
        onclick(editButton, global.editShow, id)
    }

    global.replyShow = id => {
        if (id in shownReply && shownReply[id]) {
            return;
        }

        const text = $(ID_TEXT + id);
        insertAfter(text, textareaCreate(id));
        shownReply[id] = true;

        let replyButton = $(ID_REPLY + id);

        classRemove(replyButton, 'option-reply');
        classAdd(replyButton, 'option-cancel');

        replyButton.title = 'Cancel reply';

        replyButton = removeAllEventListeners(replyButton);
        onclick(replyButton, global.replyCollapse, id);
    };

    global.replyCollapse = id => {
        let replyButton = $(ID_REPLY + id);
        const el = $(ID_SUPER_CONTAINER + id);

        el.remove();
        delete shownReply[id];

        classAdd(replyButton, 'option-reply');
        classRemove(replyButton, 'option-cancel');

        replyButton.title = 'Reply to this comment';

        replyButton = removeAllEventListeners(replyButton);
        onclick(replyButton, global.replyShow, id)
    }

    global.commentCollapse = id => {
        const children = $(ID_CHILDREN + id);
        let button = $(ID_COLLAPSE + id);

        if (children) {
            classAdd(children, 'hidden');
        }

        classRemove(button, 'option-collapse');
        classAdd(button, 'option-uncollapse');

        button.title = 'Expand children';

        button = removeAllEventListeners(button);
        onclick(button, global.commentUncollapse, id);
    }

    global.commentUncollapse = id => {
        const children = $(ID_CHILDREN + id);
        let button = $(ID_COLLAPSE + id);

        if (children) {
            classRemove(children, 'hidden');
        }

        classRemove(button, 'option-uncollapse');
        classAdd(button, 'option-collapse');

        button.title = 'Collapse children';

        button = removeAllEventListeners(button);
        onclick(button, global.commentCollapse, id);
    }

    function parentMap(comments) {
        const m = {};
        comments.forEach(comment => {
            const parentHex = comment.parentHex;
            if (!(parentHex in m)) {
                m[parentHex] = [];
            }

            comment.creationDate = new Date(comment.creationDate);

            m[parentHex].push(comment);
            commentsMap[comment.commentHex] = {
                html: comment.html,
                markdown: comment.markdown,
            };
        });

        return m;
    }

    function commentsRender() {
        const commentsArea = $(ID_COMMENTS_AREA);
        commentsArea.innerHTML = ''

        const cards = commentsRecurse(parentMap(comments), 'root');
        if (cards) {
            append(commentsArea, cards);
        }
    }

    function submitAuthenticated(id) {
        if (isAuthenticated) {
            global.commentNew(id, commenterTokenGet(), true);
            return;
        }

        global.loginBoxShow(id);
    }

    function submitAnonymous(id) {
        chosenAnonymous = true;
        global.commentNew(id, 'anonymous', true);
    }

    function submitAccountDecide(id) {
        if (requireIdentification) {
            submitAuthenticated(id);
            return;
        }

        const anonymousCheckbox = $(ID_ANONYMOUS_CHECKBOX + id);
        const textarea = $(ID_TEXTAREA + id);
        const markdown = textarea.value;

        if (markdown === '') {
            classAdd(textarea, 'red-border');
            return;
        } else {
            classRemove(textarea, 'red-border');
        }

        if (!anonymousCheckbox.checked) {
            submitAuthenticated(id);
        } else {
            submitAnonymous(id);
        }
    }

    // OAuth logic
    global.commentoAuth = data => {
        const provider = data.provider;
        const id = data.id;
        const popup = window.open('', '_blank');

        get(`${origin}/api/commenter/token/new`, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return;
            }
            errorHide();

            cookieSet('commentoCommenterToken', resp.commenterToken);

            popup.location = `${origin}/api/oauth/${provider}/redirect?commenterToken=${resp.commenterToken}`;

            const interval = setInterval(() => {
                if (popup.closed) {
                    clearInterval(interval);
                    selfGet(() => {
                        const loggedContainer = $(ID_LOGGED_CONTAINER);
                        if (loggedContainer) {
                            setAttr(loggedContainer, {style: ''});
                        }

                        if (commenterTokenGet() !== 'anonymous') {
                            remove($(ID_LOGIN));
                        }

                        if (id !== null) {
                            global.commentNew(id, resp.commenterToken, false, () => {
                                global.loginBoxClose();
                                commentsGet(commentsRender);
                            });
                        } else {
                            global.loginBoxClose();
                            commentsGet(commentsRender);
                        }
                    });
                }
            }, 250);
        });
    }

    function refreshAll(callback) {
        $(ID_ROOT).innerHTML = '';
        shownReply = {};
        global.main(callback);
    }

    function loginBoxCreate() {
        append(root, create('div', ID_LOGIN_BOX_CONTAINER));
    }

    global.popupRender = id => {
        const loginBoxContainer     = $(ID_LOGIN_BOX_CONTAINER);
        const loginBox              = create('div',    ID_LOGIN_BOX);
        const ssoSubtitle           = create('div',    ID_LOGIN_BOX_SSO_PRETEXT);
        const ssoButtonContainer    = create('div',    ID_LOGIN_BOX_SSO_BUTTON_CONTAINER);
        const ssoButton             = create('div');
        const hr1                   = create('hr',     ID_LOGIN_BOX_HR1);
        const oauthSubtitle         = create('div',    ID_LOGIN_BOX_OAUTH_PRETEXT);
        const oauthButtonsContainer = create('div',    ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER);
        const oauthButtons          = create('div');
        const hr2                   = create('hr',     ID_LOGIN_BOX_HR2);
        const emailSubtitle         = create('div',    ID_LOGIN_BOX_EMAIL_SUBTITLE);
        const emailContainer        = create('div');
        const email                 = create('div');
        const emailInput            = create('input',  ID_LOGIN_BOX_EMAIL_INPUT);
        const emailButton           = create('button', ID_LOGIN_BOX_EMAIL_BUTTON);
        const forgotLinkContainer   = create('div',    ID_LOGIN_BOX_FORGOT_LINK_CONTAINER);
        const forgotLink            = create('a');
        const loginLinkContainer    = create('div',    ID_LOGIN_BOX_LOGIN_LINK_CONTAINER);
        const loginLink             = create('a');
        const close                 = create('div');

        classAdd(loginBoxContainer, 'login-box-container');
        classAdd(loginBox, 'login-box');
        classAdd(emailSubtitle, 'login-box-subtitle');
        classAdd(emailContainer, 'email-container');
        classAdd(email, 'email');
        classAdd(emailInput, 'input');
        classAdd(emailButton, 'email-button');
        classAdd(forgotLinkContainer, 'forgot-link-container');
        classAdd(forgotLink, 'forgot-link');
        classAdd(loginLinkContainer, 'login-link-container');
        classAdd(loginLink, 'login-link');
        classAdd(ssoSubtitle, 'login-box-subtitle');
        classAdd(ssoButtonContainer, 'oauth-buttons-container');
        classAdd(ssoButton, 'oauth-buttons');
        classAdd(oauthSubtitle, 'login-box-subtitle');
        classAdd(oauthButtonsContainer, 'oauth-buttons-container');
        classAdd(oauthButtons, 'oauth-buttons');
        classAdd(close, 'login-box-close');
        classAdd(root, 'root-min-height');

        forgotLink.innerText = 'Forgot your password?';
        loginLink.innerText = 'Don\'t have an account? Sign up.';
        emailSubtitle.innerText = 'Login with your email address';
        emailButton.innerText = 'Continue';
        oauthSubtitle.innerText = 'Proceed with social login';
        ssoSubtitle.innerText = `Proceed with ${parent.location.host} authentication`;

        onclick(emailButton, global.passwordAsk, id);
        onclick(forgotLink, global.forgotPassword, id);
        onclick(loginLink, global.popupSwitch, id);
        onclick(close, global.loginBoxClose);

        setAttr(loginBoxContainer, {style: 'display: none; opacity: 0;'});
        setAttr(emailInput, {name: 'email', placeholder: 'Email address', type: 'text', autocomplete: 'email'});

        let numOauthConfigured = 0;
        const oauthProviders = ['google', 'github', 'gitlab'];
        oauthProviders.forEach(provider => {
            if (configuredOauths[provider]) {
                const button = create('button');

                classAdd(button, 'button');
                classAdd(button, `${provider}-button`);

                button.innerText = provider;

                onclick(button, global.commentoAuth, {provider: provider, id: id});

                append(oauthButtons, button);
                numOauthConfigured++;
            }
        });

        if (configuredOauths['sso']) {
            const button = create('button');

            classAdd(button, 'button');
            classAdd(button, 'sso-button');

            button.innerText = 'Single Sign-On';

            onclick(button, global.commentoAuth, {provider: 'sso', id: id});

            append(ssoButton, button);
            append(ssoButtonContainer, ssoButton);
            append(loginBox, ssoSubtitle);
            append(loginBox, ssoButtonContainer);

            if (numOauthConfigured > 0 || configuredOauths['commento']) {
                append(loginBox, hr1);
            }
        }

        if (numOauthConfigured > 0) {
            append(loginBox, oauthSubtitle);
            append(oauthButtonsContainer, oauthButtons);
            append(loginBox, oauthButtonsContainer);
            oauthButtonsShown = true;
        } else {
            oauthButtonsShown = false;
        }

        append(email, emailInput);
        append(email, emailButton);
        append(emailContainer, email);

        append(forgotLinkContainer, forgotLink);

        append(loginLinkContainer, loginLink);

        if (numOauthConfigured > 0 && configuredOauths['commento']) {
            append(loginBox, hr2);
        }

        if (configuredOauths['commento']) {
            append(loginBox, emailSubtitle);
            append(loginBox, emailContainer);
            append(loginBox, forgotLinkContainer);
            append(loginBox, loginLinkContainer);
        }

        append(loginBox, close);

        popupBoxType = 'login';
        loginBoxContainer.innerHTML = '';
        append(loginBoxContainer, loginBox);
    }

    global.forgotPassword = () => {
        const popup = window.open('', '_blank');
        popup.location = `${origin}/forgot?commenter=true`;
        global.loginBoxClose();
    }

    global.popupSwitch = id => {
        const emailSubtitle = $(ID_LOGIN_BOX_EMAIL_SUBTITLE);

        if (oauthButtonsShown) {
            remove($(ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER));
            remove($(ID_LOGIN_BOX_OAUTH_PRETEXT));
            remove($(ID_LOGIN_BOX_HR1));
            remove($(ID_LOGIN_BOX_HR2));
        }

        if (configuredOauths['sso']) {
            remove($(ID_LOGIN_BOX_SSO_BUTTON_CONTAINER));
            remove($(ID_LOGIN_BOX_SSO_PRETEXT));
            remove($(ID_LOGIN_BOX_HR1));
            remove($(ID_LOGIN_BOX_HR2));
        }

        remove($(ID_LOGIN_BOX_LOGIN_LINK_CONTAINER));
        remove($(ID_LOGIN_BOX_FORGOT_LINK_CONTAINER));

        emailSubtitle.innerText = 'Create an account';
        popupBoxType = 'signup';
        global.passwordAsk(id);
        $(ID_LOGIN_BOX_EMAIL_INPUT).focus();
    }

    function loginUP(username, password, id) {
        const json = {
            email: username,
            password: password,
        };

        post(`${origin}/api/commenter/login`, json, resp => {
            if (!resp.success) {
                global.loginBoxClose();
                errorShow(resp.message);
                return
            } else {
                errorHide();
            }

            cookieSet('commentoCommenterToken', resp.commenterToken);

            selfLoad(resp.commenter, resp.email);
            allShow();

            remove($(ID_LOGIN));
            if (id !== null) {
                global.commentNew(id, resp.commenterToken, false, () => {
                    global.loginBoxClose();
                    commentsGet(commentsRender);
                });
            } else {
                global.loginBoxClose();
                commentsGet(commentsRender);
            }
        });
    }

    global.login = id => {
        const email = $(ID_LOGIN_BOX_EMAIL_INPUT);
        const password = $(ID_LOGIN_BOX_PASSWORD_INPUT);
        loginUP(email.value, password.value, id);
    }

    global.signup = id => {
        const email = $(ID_LOGIN_BOX_EMAIL_INPUT);
        const name = $(ID_LOGIN_BOX_NAME_INPUT);
        const website = $(ID_LOGIN_BOX_WEBSITE_INPUT);
        const password = $(ID_LOGIN_BOX_PASSWORD_INPUT);

        const json = {
            email: email.value,
            name: name.value,
            website: website.value,
            password: password.value,
        };

        post(`${origin}/api/commenter/new`, json, resp => {
            if (!resp.success) {
                global.loginBoxClose();
                errorShow(resp.message);
                return
            }

            errorHide();
            loginUP(email.value, password.value, id);
        });
    }

    global.passwordAsk = id => {
        const isSignup = popupBoxType === 'signup';
        const loginBox = $(ID_LOGIN_BOX);
        const subtitle = $(ID_LOGIN_BOX_EMAIL_SUBTITLE);

        remove($(ID_LOGIN_BOX_EMAIL_BUTTON));
        remove($(ID_LOGIN_BOX_LOGIN_LINK_CONTAINER));
        remove($(ID_LOGIN_BOX_FORGOT_LINK_CONTAINER));
        if (oauthButtonsShown) {
            if (configuredOauths.length > 0) {
                remove($(ID_LOGIN_BOX_HR1));
                remove($(ID_LOGIN_BOX_HR2));
                remove($(ID_LOGIN_BOX_OAUTH_PRETEXT));
                remove($(ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER));
            }
        }

        const controls = isSignup ?
            [
                {id: ID_LOGIN_BOX_NAME_INPUT,     name: 'name',     type: 'text',     placeholder: 'Real Name'},
                {id: ID_LOGIN_BOX_WEBSITE_INPUT,  name: 'website',  type: 'text',     placeholder: 'Website (Optional)'},
                {id: ID_LOGIN_BOX_PASSWORD_INPUT, name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'new-password'},
            ] :
            [
                {id: ID_LOGIN_BOX_PASSWORD_INPUT, name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'current-password'},
            ];

        subtitle.innerText = isSignup ?
            'Finish the rest of your profile to complete.' :
            'Enter your password to log in.';

        controls.forEach(c => {
            const fieldContainer = create('div');
            const field          = create('div');
            const fieldInput     = create('input', c.id);

            classAdd(fieldContainer, 'email-container');
            classAdd(field, 'email');
            classAdd(fieldInput, 'input');

            setAttr(fieldInput, c);
            append(field, fieldInput);
            append(fieldContainer, field);

            if (c.type === 'password') {
                const fieldButton = create('button');
                classAdd(fieldButton, 'email-button');
                fieldButton.innerText = popupBoxType;

                if (isSignup) {
                    onclick(fieldButton, global.signup, id);
                } else {
                    onclick(fieldButton, global.login, id);
                }
                append(field, fieldButton);
            }
            append(loginBox, fieldContainer);
        });

        if (isSignup) {
            $(ID_LOGIN_BOX_NAME_INPUT).focus();
        } else {
            $(ID_LOGIN_BOX_PASSWORD_INPUT).focus();
        }
    }

    function pageUpdate(callback) {
        const attributes = {
            isLocked: isLocked,
            stickyCommentHex: stickyCommentHex,
        };

        const json = {
            commenterToken: commenterTokenGet(),
            domain: parent.location.host,
            path: pageId,
            attributes: attributes,
        };

        post(`${origin}/api/page/update`, json, resp => {
            if (!resp.success) {
                errorShow(resp.message);
                return
            }

            errorHide();
            call(callback);
        });
    }

    global.threadLockToggle = () => {
        const lock = $(ID_MOD_TOOLS_LOCK_BUTTON);

        isLocked = !isLocked;

        lock.disabled = true;
        pageUpdate(() => {
            lock.disabled = false;
            refreshAll();
        });
    }

    global.commentSticky = commentHex => {
        if (stickyCommentHex !== 'none') {
            const sticky = $(ID_STICKY + stickyCommentHex);
            classRemove(sticky, 'option-unsticky');
            classAdd(sticky, 'option-sticky');
        }

        stickyCommentHex = stickyCommentHex === commentHex ? 'none' : commentHex;

        pageUpdate(() => {
            const sticky = $(ID_STICKY + commentHex);
            if (stickyCommentHex === commentHex) {
                classRemove(sticky, 'option-sticky');
                classAdd(sticky, 'option-unsticky');
            } else {
                classRemove(sticky, 'option-unsticky');
                classAdd(sticky, 'option-sticky');
            }
        });
    }

    function mainAreaCreate() {
        const mainArea = create('div', ID_MAIN_AREA);
        classAdd(mainArea, 'main-area');
        setAttr(mainArea, {style: 'display: none'});
        append(root, mainArea);
    }

    function modToolsCreate() {
        const modTools = create('div',    ID_MOD_TOOLS);
        const lock     = create('button', ID_MOD_TOOLS_LOCK_BUTTON);
        classAdd(modTools, 'mod-tools');

        lock.innerHTML = isLocked ? 'Unlock Thread' : 'Lock Thread';

        onclick(lock, global.threadLockToggle);

        setAttr(modTools, {style: 'display: none'});

        append(modTools, lock);
        append(root, modTools);
    }

    function loadCssOverride() {
        if (cssOverride === undefined) {
            allShow();
        } else {
            cssLoad(cssOverride, allShow);
        }
    }

    function allShow() {
        const mainArea = $(ID_MAIN_AREA);
        const modTools = $(ID_MOD_TOOLS);
        const loggedContainer = $(ID_LOGGED_CONTAINER);

        setAttr(mainArea, {style: ''});

        if (isModerator) {
            setAttr(modTools, {style: ''});
        }

        if (loggedContainer) {
            setAttr(loggedContainer, {style: ''});
        }
    }

    global.loginBoxClose = () => {
        const mainArea = $(ID_MAIN_AREA);
        const loginBoxContainer = $(ID_LOGIN_BOX_CONTAINER);

        classRemove(mainArea, 'blurred');
        classRemove(root, 'root-min-height');

        setAttr(loginBoxContainer, {style: 'display: none'});
    }

    global.loginBoxShow = id => {
        const mainArea = $(ID_MAIN_AREA);
        const loginBoxContainer = $(ID_LOGIN_BOX_CONTAINER);

        global.popupRender(id);

        classAdd(mainArea, 'blurred');
        setAttr(loginBoxContainer, {style: ''});

        window.location.hash = ID_LOGIN_BOX_CONTAINER;

        $(ID_LOGIN_BOX_EMAIL_INPUT).focus();
    }

    function dataTagsLoad() {
        const scripts = tags('script');
        for (let i = 0; i < scripts.length; i++) {
            const scr = scripts[i];
            if (scr.src.match(/\/js\/commento\.js$/)) {
                const pid = getAttr(scr, 'data-page-id');
                if (pid !== undefined) {
                    pageId = pid;
                }
                cssOverride = getAttr(scr, 'data-css-override');
                autoInit = getAttr(scr, 'data-auto-init');
                ID_ROOT = getAttr(scr, 'data-id-root');
                if (ID_ROOT === undefined) {
                    ID_ROOT = 'commento';
                }
                noFonts = getAttr(scr, 'data-no-fonts');
                hideDeleted = getAttr(scr, 'data-hide-deleted');
            }
        }
    }

    function loadHash() {
        if (window.location.hash) {
            if (window.location.hash.startsWith('#commento-')) {
                const id = window.location.hash.split('-')[1];
                const el = $(ID_CARD + id);
                if (el === null) {
                    if (id.length === 64) {
                        // A hack to make sure it's a valid ID before showing the user a message.
                        errorShow('The comment you\'re looking for no longer exists or was deleted.');
                    }
                    return;
                }

                classAdd(el, 'highlighted-card');
                el.scrollIntoView(true);
            } else if (window.location.hash.startsWith('#commento')) {
                root.scrollIntoView(true);
            }
        }
    }

    global.main = callback => {
        root = $(ID_ROOT);
        if (root === null) {
            console.log(`[commento] error: no root element with ID '${ID_ROOT}' found`);
            return;
        }

        if (mobileView === null) {
            mobileView = root.getBoundingClientRect()['width'] < 450;
        }

        classAdd(root, 'root');
        if (noFonts !== 'true') {
            classAdd(root, 'root-font');
        }

        loginBoxCreate();

        errorElementCreate();

        mainAreaCreate();

        const footer = footerLoad();
        cssLoad(`${cdn}/css/commento.css`, loadCssOverride);

        selfGet(() => {
            commentsGet(() => {
                modToolsCreate();
                rootCreate(() => {
                    commentsRender();
                    append(root, footer);
                    loadHash();
                    allShow();
                    nameWidthFix();
                    call(callback);
                });
            });
        });
    }

    let initted = false;

    function init() {
        if (initted) {
            return;
        }
        initted = true;

        dataTagsLoad();

        if (autoInit === 'true' || autoInit === undefined) {
            global.main(undefined);
        } else if (autoInit !== 'false') {
            console.log('[commento] error: invalid value for data-auto-init; allowed values: true, false');
        }
    }

    const readyLoad = () => {
        switch (document.readyState) {
        // The document is still loading. The div we need to fill might not have
        // been parsed yet, so let's wait and retry when the readyState changes.
        // If there is more than one state change, we aren't affected because we
        // have a double-call protection in init().
        case 'loading':
            document.addEventListener('readystatechange', readyLoad);
            break;

        // The document has been parsed and DOM objects are now accessible. While
        // JS, CSS, and images are still loading, we don't need to wait.
        case 'interactive':
            init();
            break;

        // The page has fully loaded (including JS, CSS, and images). From our
        // point of view, this is practically no different from interactive.
        case 'complete':
            init();
            break;
        }
    };

    readyLoad();

}(window.commento, document));
