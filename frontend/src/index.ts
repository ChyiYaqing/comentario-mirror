// noinspection DuplicatedCode

import { Comment, Commenter, CommentMap, CommentsMap, Email, OAuthResponse, SortPolicy } from './models';
import { HttpClient } from './http-client';
import {
    ApiCommentEditResponse,
    ApiCommenterLoginResponse,
    ApiCommenterTokenNewResponse,
    ApiCommentListResponse,
    ApiCommentNewResponse,
    ApiResponseBase,
    ApiSelfResponse,
} from './api';

type CallbackFunction = () => void;

interface ElementConfig {
    id?:        string;
    classes?:   string | string[];
    innerText?: string;
    innerHTML?: string;
    parent?:    HTMLElement;
    children?:  HTMLElement[];

    [ k: string ]: any;
}

(function (global, document) {

    const origin = '[[[.Origin]]]';
    const cdn = '[[[.CdnPrefix]]]';

    /** HTTP client we'll use for API requests. */
    const apiClient = new HttpClient(`${origin}/api`);

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
    const ID_LOGIN_BOX_PASSWORD_BUTTON = 'commento-login-box-password-button';
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

    let root: HTMLElement = null;
    let pageId = parent.location.pathname;
    let cssOverride: string;
    let noFonts = false;
    let hideDeleted = false;
    let autoInit = true;
    let isAuthenticated = false;
    let comments: Comment[] = [];
    const commentsByHex: CommentMap = {};
    let commenters: { [k: string]: Commenter } = {};
    let requireIdentification = true;
    let isModerator = false;
    let isFrozen = false;
    let chosenAnonymous = false;
    let isLocked = false;
    let stickyCommentHex = 'none';
    let shownReply: { [k: string]: boolean } = {};
    const shownEdit: { [k: string]: boolean } = {};
    let configuredOauths: { [k: string]: boolean } = {};
    let anonymousOnly = false;
    let popupBoxType = 'login';
    let oauthButtonsShown = false;
    let sortPolicy: SortPolicy = 'score-desc';
    let selfHex: string = undefined;
    let mobileView: boolean | null = null;
    const loadedCss: { [k: string]: boolean } = {};

    function byId<T extends HTMLElement>(id: string): T {
        return document.getElementById(id) as T;
    }

    function prepend(root: HTMLElement, el: HTMLElement) {
        root.prepend(el);
    }

    function append(parent: HTMLElement, ...children: HTMLElement[]) {
        children.forEach(c => parent.appendChild(c));
    }

    function insertAfter(el1: HTMLElement, el2: HTMLElement) {
        el1.parentNode.insertBefore(el2, el1.nextSibling);
    }

    /**
     * Add the provided class or classes to the element.
     * @param el Element to add classes to.
     * @param classes string|array Class(es) to add. Falsy values are ignored.
     */
    function addClasses(el: HTMLElement, classes: string | string[]) {
        (Array.isArray(classes) ? classes : [classes]).forEach(c => c && el.classList.add(`commento-${c}`));
    }

    /**
     * Remove the provided class or classes from the element.
     * @param el Element to remove classes from.
     * @param classes string|array Class(es) to remove. Falsy values are ignored.
     */
    function removeClasses(el: HTMLElement, classes: string | string[]) {
        if (el !== null) {
            (Array.isArray(classes) ? classes : [classes]).forEach(c => c && el.classList.remove(`commento-${c}`));
        }
    }

    /**
     * Create a new HTML element with the given tag and configuration.
     * @param tagName Name of the tag.
     * @param config Optional configuration object.
     * @returns {*} The created and configured HTML element.
     */
    function create<K extends keyof HTMLElementTagNameMap>(tagName: K, config?: ElementConfig): HTMLElementTagNameMap[K] {
        // Create a new HTML element
        const e = document.createElement(tagName) as HTMLElementTagNameMap[K];

        // If there's any config passed
        if (config) {
            // Set up the ID, if given, and clean it up from the config
            if ('id' in config) {
                e.id = config.id;
                delete config.id;
            }

            // Set up the classes, if given, and clean them up from the config
            if ('classes' in config) {
                addClasses(e, config.classes);
                delete config.classes;
            }

            // Set up the inner text/HTML, if given, and clean it up from the config
            if ('innerText' in config) {
                e.innerText = config.innerText;
                delete config.innerText;
            } else if ('innerHTML' in config) {
                e.innerHTML = config.innerHTML;
                delete config.innerHTML;
            }

            // Set up the parent, if given, and clean it up from the config
            let parent: HTMLElement;
            if ('parent' in config) {
                parent = config.parent;
                delete config.parent;
            }

            // Add any children
            if ('children' in config) {
                config.children.forEach(child => e.appendChild(child));
                delete config.children;
            }

            // Set up the remaining attributes
            setAttr(e, config)

            // Add the child to the parent, if any
            if (parent) {
                parent.appendChild(e)
            }
        }
        return e;
    }

    function remove(...elements: HTMLElement[]) {
        elements?.forEach(e => e && e.parentNode.removeChild(e));
    }

    function getAttr(node: HTMLElement, attrName: string) {
        const attr = node.attributes.getNamedItem(attrName);
        return attr === undefined ? undefined : attr?.value;
    }

    function removeAllEventListeners<T extends HTMLElement>(node: T): T {
        if (node) {
            const replacement = node.cloneNode(true) as T;
            if (node.parentNode !== null) {
                node.parentNode.replaceChild(replacement, node);
                return replacement;
            }
        }
        return node;
    }

    /**
     * Bind a handler to the onClick event of the given element.
     * @param e Element to bind a handler to.
     * @param handler Handler to bind.
     */
    function onClick(e: HTMLElement, handler: CallbackFunction) {
        e.addEventListener('click', handler, false);
    }

    /**
     * Set node attributes from the provided object.
     * @param node HTML element to set attributes on.
     * @param values Object that provides attribute names (keys) and their values. null and undefined values cause attribute removal from the node.
     */
    function setAttr(node: HTMLElement, values: { [k: string]: string }) {
        if (node) {
            Object.keys(values).forEach(k => {
                const v = values[k];
                if (v === undefined || v === null) {
                    node.removeAttribute(k);
                } else {
                    node.setAttribute(k, v)
                }
            });
        }
    }

    function cookieGet(name: string): string {
        const c = `; ${document.cookie}`;
        const x = c.split(`; ${name}=`);
        return x.length === 2 ? x.pop().split(';').shift() : null;
    }

    function cookieSet(name: string, value: string) {
        const date = new Date();
        date.setTime(date.getTime() + (365 * 24 * 60 * 60 * 1000));
        document.cookie = `${name}=${value}; expires=${date.toUTCString()}; path=/`;
    }

    function commenterTokenGet() {
        const commenterToken = cookieGet('commentoCommenterToken');
        return commenterToken === undefined ? 'anonymous' : commenterToken;
    }

    global.logout = (): Promise<void> => {
        cookieSet('commentoCommenterToken', 'anonymous');
        isAuthenticated = false;
        isModerator = false;
        selfHex = undefined;
        return refreshAll();
    }

    function profileEdit() {
        window.open(`${origin}/profile?commenterToken=${commenterTokenGet()}`, '_blank');
    }

    function notificationSettings(unsubscribeSecretHex: string) {
        window.open(`${origin}/unsubscribe?unsubscribeSecretHex=${unsubscribeSecretHex}`, '_blank');
    }

    function selfLoad(commenter: Commenter, email: Email) {
        commenters[commenter.commenterHex] = commenter;
        selfHex = commenter.commenterHex;

        const loggedContainer = create('div', {id: ID_LOGGED_CONTAINER, classes: 'logged-container', style: 'display: none'});
        const loggedInAs      = create('div', {classes: 'logged-in-as', parent: loggedContainer});
        const name            = create(commenter.link !== 'undefined' ? 'a' : 'div', {classes: 'name', innerText: commenter.name, parent: loggedInAs});
        const btnSettings     = create('div', {classes: 'profile-button', innerText: 'Notification Settings'});
        const btnEditProfile  = create('div', {classes: 'profile-button', innerText: 'Edit Profile'});
        const btnLogout       = create('div', {classes: 'profile-button', innerText: 'Logout', parent: loggedContainer});
        const color = colorGet(`${commenter.commenterHex}-${commenter.name}`);

        // Set the profile href for the commenter, if any
        if (commenter.link !== 'undefined') {
            setAttr(name, {href: commenter.link});
        }

        onClick(btnLogout,      () => global.logout());
        onClick(btnSettings,    () => notificationSettings(email.unsubscribeSecretHex));
        onClick(btnEditProfile, () => profileEdit);

        // Add an avatar
        if (commenter.photo === 'undefined') {
            create('div', {
                classes:   'avatar',
                innerHTML: commenter.name[0].toUpperCase(),
                style:     `background-color: ${color}`,
                parent:    loggedInAs,
            });
        } else {
            create('img', {
                classes: 'avatar-img',
                src:     `${cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                loading: 'lazy',
                alt:     '',
                parent:  loggedInAs,
            });
        }

        // If it's a local user, add an Edit profile button
        if (commenter.provider === 'commento') {
            append(loggedContainer, btnEditProfile);
        }
        append(loggedContainer, btnSettings);

        // Add the container to the root
        prepend(root, loggedContainer);
        isAuthenticated = true;
    }

    function selfGet(): Promise<void> {
        const commenterToken = commenterTokenGet();
        if (commenterToken === 'anonymous') {
            isAuthenticated = false;
            return Promise.resolve();
        }

        return apiClient.post<ApiSelfResponse>('commenter/self', {commenterToken: commenterTokenGet()})
            .then(resp => {
                if (!resp.success) {
                    cookieSet('commentoCommenterToken', 'anonymous');
                    return Promise.reject(resp.message);
                }

                selfLoad(resp.commenter, resp.email);
                allShow();
                return undefined;
            });
    }

    /**
     * Load the stylesheet with the provided URL into the DOM
     * @param url Stylesheet URL.
     */
    function cssLoad(url: string): Promise<void> {
        // Don't bother if the stylesheet has been loaded already
        return loadedCss[url] ?
            Promise.resolve() :
            new Promise(resolve => {
                loadedCss[url] = true;
                const link = create('link', {href: url, rel: 'stylesheet', type: 'text/css'});
                link.addEventListener('load', () => resolve());
                append(document.getElementsByTagName('head')[0], link);
            });
    }

    function footerLoad() {
        return create('div', {
            id:       ID_FOOTER,
            classes:  'footer',
            children: [
                create('div', {
                    classes:  'logo-container',
                    children: [
                        create('a', {
                            classes:  'logo',
                            href:     'https://comentario.app/',
                            target:   '_blank',
                            children: [
                                create('span', {classes: 'logo-text', innerText: 'Comentario 🗨'}),
                            ],
                        }),
                    ],
                }),
            ],
        });
    }

    function commentsGet(): Promise<void> {
        return apiClient.post<ApiCommentListResponse>(
            'comment/list',
            {
                commenterToken: commenterTokenGet(),
                domain:         parent.location.host,
                path:           pageId,
            })
            .then(resp => {
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
                commenters = Object.assign({}, commenters, resp.commenters);
                configuredOauths = resp.configuredOauths;

                sortPolicy = resp.defaultSortPolicy;
            });
    }

    function errorShow(text: string) {
        const el = byId<HTMLDivElement>(ID_ERROR);
        el.innerText = text;
        setAttr(el, {style: 'display: block;'});
    }

    function errorHide() {
        setAttr(byId(ID_ERROR), {style: 'display: none;'});
    }

    function errorElementCreate() {
        create('div', {id: ID_ERROR, classes: 'error-box', style: 'display: none;', parent: root});
    }

    function autoExpander(el: HTMLElement): CallbackFunction {
        return () => {
            el.style.height = '';
            el.style.height = `${Math.min(Math.max(el.scrollHeight, 75), 400)}px`;
        }
    }

    function markdownHelpShow(id: string) {
        create('table', {
            id:       ID_MARKDOWN_HELP + id,
            classes:  'markdown-help',
            parent:   byId(ID_SUPER_CONTAINER + id),
            children: [
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<i>italics</i>'}),
                        create('td', {innerHTML: 'surround text with <pre>*asterisks*</pre>'}),
                    ],
                }),
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<b>bold</b>'}),
                        create('td', {innerHTML: 'surround text with <pre>**two asterisks**</pre>'}),
                    ],
                }),
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<pre>code</pre>'}),
                        create('td', {innerHTML: 'surround text with <pre>`backticks`</pre>'}),
                    ],
                }),
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<del>strikethrough</del>'}),
                        create('td', {innerHTML: 'surround text with <pre>~~two tilde characters~~</pre>'}),
                    ],
                }),
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<a href="https://example.com">hyperlink</a>'}),
                        create('td', {innerHTML: '<pre>[hyperlink](https://example.com)</pre> or just a bare URL'}),
                    ],
                }),
                create('tr', {
                    children: [
                        create('td', {innerHTML: '<blockquote>quote</blockquote>'}),
                        create('td', {innerHTML: 'prefix with <pre>&gt;</pre>'}),
                    ],
                }),
            ],
        });

        // Add a collapse button
        const markdownButton = removeAllEventListeners(byId<HTMLAnchorElement>(ID_MARKDOWN_BUTTON + id));
        onClick(markdownButton, () => markdownHelpHide(id));
    }

    function markdownHelpHide(id: string) {
        let markdownButton = byId<HTMLAnchorElement>(ID_MARKDOWN_BUTTON + id);
        const markdownHelp = byId(ID_MARKDOWN_HELP + id);

        markdownButton = removeAllEventListeners(markdownButton);
        onClick(markdownButton, () => markdownHelpShow(id));
        remove(markdownHelp);
    }

    function textareaCreate(id: string, edit: boolean) {
        const textOuter        = create('div',      {id: ID_SUPER_CONTAINER + id, classes: 'button-margin'});
        const textCont         = create('div',      {id: ID_TEXTAREA_CONTAINER + id, classes: 'textarea-container', parent: textOuter});
        const textArea         = create('textarea', {id: ID_TEXTAREA + id, placeholder: 'Add a comment', parent: textCont});
        const anonCheckbox     = create('input',    {id: ID_ANONYMOUS_CHECKBOX + id, type: 'checkbox'});
        const anonCheckboxCont = create('div', {
            classes:  ['round-check', 'anonymous-checkbox-container'],
            children: [
                anonCheckbox,
                create('label', {for: ID_ANONYMOUS_CHECKBOX + id, innerText: 'Comment anonymously'}),
            ],
        });
        const submitButton = create('button', {
            id:        ID_SUBMIT_BUTTON + id,
            classes:   ['button', 'submit-button'],
            innerText: edit ? 'Save Changes' : 'Add Comment',
            parent:    textOuter,
        });
        const markdownButton = create('a', {
            id:        ID_MARKDOWN_BUTTON + id,
            classes:   'markdown-button',
            innerHTML: '<b>M⬇</b>&nbsp;Markdown',
        });

        if (anonymousOnly) {
            anonCheckbox.checked = true;
            anonCheckbox.setAttribute('disabled', 'true');
        }

        textArea.oninput = autoExpander(textArea);
        onClick(submitButton, () => edit ? commentEdit(id) : submitAccountDecide(id));
        onClick(markdownButton, () => markdownHelpShow(id));
        if (!requireIdentification && !edit) {
            append(textOuter, anonCheckboxCont);
        }
        append(textOuter, markdownButton);
        return textOuter;
    }

    const sortPolicyNames: { [k in SortPolicy]: string } = {
        'score-desc':        'Upvotes',
        'creationdate-desc': 'Newest',
        'creationdate-asc':  'Oldest',
    };

    function sortPolicyApply(policy: SortPolicy) {
        removeClasses(byId(ID_SORT_POLICY + sortPolicy), 'sort-policy-button-selected');

        const commentsArea = byId<HTMLDivElement>(ID_COMMENTS_AREA);
        commentsArea.innerHTML = '';
        sortPolicy = policy;
        const cards = commentsRecurse(parentMap(comments), 'root');
        if (cards) {
            append(commentsArea, cards);
        }

        addClasses(byId(ID_SORT_POLICY + policy), 'sort-policy-button-selected');
    }

    function sortPolicyBox(): HTMLDivElement {
        const container = create('div', {classes: 'sort-policy-buttons-container'});
        const buttonBar = create('div', {classes: 'sort-policy-buttons', parent: container});
        Object.keys(sortPolicyNames).forEach((sp: SortPolicy) => {
            const sortPolicyButton = create('a', {
                id:        ID_SORT_POLICY + sp,
                classes:   ['sort-policy-button', sp === sortPolicy && 'sort-policy-button-selected'],
                innerText: sortPolicyNames[sp],
                parent:    buttonBar,
            });
            onClick(sortPolicyButton, () => sortPolicyApply(sp));
        });
        return container;
    }

    function rootCreate(): void {
        const mainArea = byId(ID_MAIN_AREA);
        const login           = create('div', {id: ID_LOGIN, classes: 'login'});
        const loginText       = create('div', {classes: 'login-text', innerText: 'Login'});
        const preCommentsArea = create('div', {id: ID_PRE_COMMENTS_AREA});
        const commentsArea    = create('div', {id: ID_COMMENTS_AREA, classes: 'comments'});
        onClick(loginText, () => global.loginBoxShow(null));

        // If there's an OAuth provider configured, add a Login button
        if (Object.keys(configuredOauths).some(k => configuredOauths[k])) {
            append(login, loginText);
        } else if (!requireIdentification) {
            anonymousOnly = true;
        }

        if (isLocked || isFrozen) {
            if (isAuthenticated || chosenAnonymous) {
                append(mainArea, messageCreate('This thread is locked. You cannot add new comments.'));
                remove(byId(ID_LOGIN));
            } else {
                append(mainArea, login, textareaCreate('root', false));
            }
        } else {
            if (isAuthenticated) {
                remove(byId(ID_LOGIN));
            } else {
                append(mainArea, login);
            }
            append(mainArea, textareaCreate('root', false));
        }

        if (comments.length > 0) {
            append(mainArea, sortPolicyBox());
        }
        append(mainArea, preCommentsArea, commentsArea);
        append(root, mainArea);
    }

    function messageCreate(text: string): HTMLDivElement {
        return create('div', {classes: 'moderation-notice', innerText: text});
    }

    global.commentNew = (id: string, commenterToken: string, appendCard: boolean): Promise<void> => {
        const container   = byId<HTMLDivElement>(ID_SUPER_CONTAINER + id);
        const textarea    = byId<HTMLTextAreaElement>(ID_TEXTAREA + id);
        const replyButton = byId<HTMLButtonElement>(ID_REPLY + id);

        const markdown = textarea.value;

        if (markdown === '') {
            addClasses(textarea, 'red-border');
            return Promise.reject();
        }

        removeClasses(textarea, 'red-border');

        const data = {
            commenterToken,
            domain: parent.location.host,
            path: pageId,
            parentHex: id,
            markdown,
        };

        return apiClient.post<ApiCommentNewResponse>('comment/new', data)
            .then(resp => {
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
                    prepend(byId(ID_SUPER_CONTAINER + id), messageCreate(message));
                }

                const comment: Comment = {
                    commentHex: resp.commentHex,
                    commenterHex: selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : selfHex,
                    markdown,
                    html: resp.html,
                    parentHex: 'root',
                    score: 0,
                    state: 'approved',
                    direction: 0,
                    creationDate: new Date().toISOString(),
                };

                const newCard = commentsRecurse({root: [comment]}, 'root');

                commentsByHex[resp.commentHex] = comment;
                if (appendCard) {
                    if (id !== 'root') {
                        container.replaceWith(newCard);

                        shownReply[id] = false;

                        addClasses(replyButton, 'option-reply');
                        removeClasses(replyButton, 'option-cancel');

                        replyButton.title = 'Reply to this comment';

                        onClick(replyButton, () => global.replyShow(id));
                    } else {
                        textarea.value = '';
                        insertAfter(byId(ID_PRE_COMMENTS_AREA), newCard);
                    }
                } else if (id === 'root') {
                    textarea.value = '';
                }
            });
    }

    function colorGet(name: string) {
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

    function timeDifference(current: number, previous: number) {
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

    function scorify(score: number) {
        return score === 1 ? 'One point' : `${score} points`;
    }

    const sortPolicyFunctions: { [k in SortPolicy]: (a: Comment, b: Comment) => number } = {
        'score-desc':        (a, b) => b.score - a.score,
        'creationdate-desc': (a, b) => a.creationMs < b.creationMs ? 1 : -1,
        'creationdate-asc':  (a, b) => a.creationMs < b.creationMs ? -1 : 1,
    };

    function commentsRecurse(parentMap: CommentsMap, parentHex: string) {
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
            const hex = comment.commentHex;
            const header = create('div', {classes: 'header'});
            const name = create(
                commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '' ? 'a' : 'div',
                {
                    id:        ID_NAME + hex,
                    innerText: comment.deleted ? '[deleted]' : commenter.name,
                    classes:   'name',
                });
            const color = colorGet(`${comment.commenterHex}-${commenter.name}`);
            const card     = create('div', {id: ID_CARD     + hex, style: `border-left: 2px solid ${color}`, classes: 'card'});
            const subtitle = create('div', {id: ID_SUBTITLE + hex, classes: 'subtitle'});
            const timeago = create('div', {
                id:        ID_TIMEAGO + hex,
                classes:   'timeago',
                innerHTML: timeDifference(curTime, comment.creationMs),
                title:     comment.creationDate.toString(),
            });
            const score = create('div', {id: ID_SCORE + hex, classes: 'score', innerText: scorify(comment.score)});
            const body     = create('div',    {id: ID_BODY     + hex, classes: 'body'});
            const text     = create('div',    {id: ID_TEXT     + hex, innerHTML: comment.html});
            const options  = create('div',    {id: ID_OPTIONS  + hex, classes: 'options'});
            const edit     = create('button', {id: ID_EDIT     + hex, classes: ['option-button', 'option-edit'],     title: 'Edit'});
            const reply    = create('button', {id: ID_REPLY    + hex, classes: ['option-button', 'option-reply'],    title: 'Reply'});
            const collapse = create('button', {id: ID_COLLAPSE + hex, classes: ['option-button', 'option-collapse'], title: 'Collapse children'});
            let   upvote   = create('button', {id: ID_UPVOTE   + hex, classes: ['option-button', 'option-upvote'],   title: 'Upvote'});
            let   downvote = create('button', {id: ID_DOWNVOTE + hex, classes: ['option-button', 'option-downvote'], title: 'Downvote'});
            const approve  = create('button', {id: ID_APPROVE  + hex, classes: ['option-button', 'option-approve'],  title: 'Approve'});
            const remove   = create('button', {id: ID_REMOVE   + hex, classes: ['option-button', 'option-remove'],   title: 'Remove'});
            const sticky   = create('button', {
                id:      ID_STICKY + hex,
                classes: ['option-button', stickyCommentHex === hex ? 'option-unsticky' : 'option-sticky'],
                title:   stickyCommentHex === hex ? isModerator ? 'Unsticky' : 'This comment has been stickied' : 'Sticky',
            });
            const contents = create('div',    {id: ID_CONTENTS + hex});
            if (mobileView) {
                addClasses(options, 'options-mobile');
            }

            const children = commentsRecurse(parentMap, hex);
            if (children) {
                children.id = ID_CHILDREN + hex;
            }

            let avatar;
            if (commenter.photo === 'undefined') {
                avatar = create('div', {style: `background-color: ${color}`, classes: 'avatar'});

                if (comment.commenterHex === 'anonymous') {
                    avatar.innerHTML = '?';
                    avatar.style.fontWeight = 'bold';
                } else {
                    avatar.innerHTML = commenter.name[0].toUpperCase();
                }
            } else {
                create('img', {
                    src:     `${cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                    classes: 'avatar-img',
                });
            }
            if (isModerator && comment.state !== 'approved') {
                addClasses(card, 'dark-card');
            }
            if (commenter.isModerator) {
                addClasses(name, 'moderator');
            }
            if (comment.state === 'flagged') {
                addClasses(name, 'flagged');
            }

            if (isAuthenticated) {
                if (comment.direction > 0) {
                    addClasses(upvote, 'upvoted');
                } else if (comment.direction < 0) {
                    addClasses(downvote, 'downvoted');
                }
            }

            onClick(edit,     () => global.editShow(hex));
            onClick(collapse, () => global.commentCollapse(hex));
            onClick(approve,  () => global.commentApprove(hex));
            onClick(remove,   () => global.commentDelete(hex));
            onClick(sticky,   () => global.commentSticky(hex));

            if (isAuthenticated) {
                const upDown = upDownOnClickSet(upvote, downvote, hex, comment.direction);
                upvote = upDown[0];
                downvote = upDown[1];
            } else {
                onClick(upvote,   () => global.loginBoxShow(null));
                onClick(downvote, () => global.loginBoxShow(null));
            }

            onClick(reply, () => global.replyShow(hex));

            if (commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '') {
                setAttr(name, {href: commenter.link});
            }

            append(options, collapse);

            if (!comment.deleted) {
                append(options, downvote, upvote);
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
                setAttr(options.children[i] as HTMLElement, {style: `right: ${i * 32}px;`});
            }

            append(subtitle, score, timeago);

            if (!mobileView) {
                append(header, options);
            }
            append(header, avatar, name, subtitle);
            append(body, text);
            append(contents, body);
            if (mobileView) {
                append(contents, options);
                create('div', {classes: 'options-clearfix', parent: contents});
            }

            if (children) {
                addClasses(children, 'body');
                append(contents, children);
            }

            append(card, header, contents);

            if (comment.deleted && (hideDeleted || children === null)) {
                return;
            }

            append(cards, card);
        });

        return cards.childNodes.length ? cards : null;
    }

    global.commentApprove = (commentHex: string): Promise<void> =>
        apiClient.post<ApiResponseBase>('comment/approve', {commenterToken: commenterTokenGet(), commentHex})
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    return;
                }
                errorHide();

                const card = byId(ID_CARD + commentHex);
                const name = byId(ID_NAME + commentHex);
                const tick = byId(ID_APPROVE + commentHex);

                removeClasses(card, 'dark-card');
                removeClasses(name, 'flagged');
                remove(tick);
            })

    global.commentDelete = (commentHex: string): Promise<void> => {
        if (!confirm('Are you sure you want to delete this comment?')) {
            return Promise.reject();
        }

        return apiClient.post<ApiResponseBase>('comment/delete', {commenterToken: commenterTokenGet(), commentHex})
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    return;
                }

                errorHide();
                const text = byId<HTMLDivElement>(ID_TEXT + commentHex);
                text.innerText = '[deleted]';
            });
    }

    function nameWidthFix() {
        const els = document.getElementsByClassName('commento-name');

        for (let i = 0; i < els.length; i++) {
            setAttr(els[i] as HTMLElement, {style: `max-width: ${els[i].getBoundingClientRect()['width'] + 20}px;`})
        }
    }

    function upDownOnClickSet(upvote: HTMLButtonElement, downvote: HTMLButtonElement, commentHex: string, direction: number): [HTMLButtonElement, HTMLButtonElement] {
        upvote = removeAllEventListeners(upvote);
        downvote = removeAllEventListeners(downvote);

        if (direction > 0) {
            onClick(upvote,   () => global.vote(commentHex, 1, 0));
            onClick(downvote, () => global.vote(commentHex, 1, -1));
        } else if (direction < 0) {
            onClick(upvote,   () => global.vote(commentHex, -1, 1));
            onClick(downvote, () => global.vote(commentHex, -1, 0));
        } else {
            onClick(upvote,   () => global.vote(commentHex, 0, 1));
            onClick(downvote, () => global.vote(commentHex, 0, -1));
        }

        return [upvote, downvote];
    }

    global.vote = (commentHex: string, oldDirection: number, direction: number): Promise<void> => {
        let upvote   = byId<HTMLButtonElement>(ID_UPVOTE + commentHex);
        let downvote = byId<HTMLButtonElement>(ID_DOWNVOTE + commentHex);
        const score  = byId<HTMLDivElement>(ID_SCORE + commentHex);

        const upDown = upDownOnClickSet(upvote, downvote, commentHex, direction);
        upvote = upDown[0];
        downvote = upDown[1];

        removeClasses(upvote, 'upvoted');
        removeClasses(downvote, 'downvoted');
        if (direction > 0) {
            addClasses(upvote, 'upvoted');
        } else if (direction < 0) {
            addClasses(downvote, 'downvoted');
        }

        score.innerText = scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) + direction - oldDirection);

        return apiClient.post<ApiResponseBase>('comment/vote', {commenterToken: commenterTokenGet(), commentHex, direction})
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    removeClasses(upvote, 'upvoted');
                    removeClasses(downvote, 'downvoted');
                    score.innerText = scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) - direction + oldDirection);
                    upDownOnClickSet(upvote, downvote, commentHex, oldDirection);
                    return;
                }
                errorHide();
            });
    }

    function commentEdit(id: string): Promise<void> {
        const textarea = byId<HTMLTextAreaElement>(ID_TEXTAREA + id);
        const markdown = textarea.value;
        if (markdown === '') {
            addClasses(textarea, 'red-border');
            return Promise.reject();
        }

        removeClasses(textarea, 'red-border');

        const data = {
            commenterToken: commenterTokenGet(),
            commentHex: id,
            markdown,
        };

        return apiClient.post<ApiCommentEditResponse>('comment/edit', data)
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    return;
                }

                errorHide();

                commentsByHex[id].markdown = markdown;
                commentsByHex[id].html = resp.html;

                let editButton = byId<HTMLButtonElement>(ID_EDIT + id);
                const textarea = byId<HTMLTextAreaElement>(ID_SUPER_CONTAINER + id);

                textarea.innerHTML = commentsByHex[id].html;
                textarea.id = ID_TEXT + id;
                delete shownEdit[id];

                addClasses(editButton, 'option-edit');
                removeClasses(editButton, 'option-cancel');

                editButton.title = 'Edit comment';

                editButton = removeAllEventListeners(editButton);
                onClick(editButton, () => global.editShow(id));

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
                    prepend(byId(ID_SUPER_CONTAINER + id), messageCreate(message));
                }
            });
    }

    global.editShow = (id: string) => {
        if (shownEdit[id]) {
            return;
        }

        const text = byId<HTMLDivElement>(ID_TEXT + id);
        shownEdit[id] = true;
        text.replaceWith(textareaCreate(id, true));

        const textarea = byId<HTMLTextAreaElement>(ID_TEXTAREA + id);
        textarea.value = commentsByHex[id].markdown;

        let editButton = byId<HTMLButtonElement>(ID_EDIT + id);

        removeClasses(editButton, 'option-edit');
        addClasses(editButton, 'option-cancel');

        editButton.title = 'Cancel edit';

        editButton = removeAllEventListeners(editButton);
        onClick(editButton, () => global.editCollapse(id));
    };

    global.editCollapse = (id: string) => {
        let editButton = byId(ID_EDIT + id);
        const textarea = byId(ID_SUPER_CONTAINER + id);

        textarea.innerHTML = commentsByHex[id].html;
        textarea.id = ID_TEXT + id;
        delete shownEdit[id];

        addClasses(editButton, 'option-edit');
        removeClasses(editButton, 'option-cancel');

        editButton.title = 'Edit comment';

        editButton = removeAllEventListeners(editButton);
        onClick(editButton, () => global.editShow(id))
    }

    global.replyShow = (id: string) => {
        if (id in shownReply && shownReply[id]) {
            return;
        }

        const text = byId(ID_TEXT + id);
        insertAfter(text, textareaCreate(id, false));
        shownReply[id] = true;

        let replyButton = byId(ID_REPLY + id);

        removeClasses(replyButton, 'option-reply');
        addClasses(replyButton, 'option-cancel');

        replyButton.title = 'Cancel reply';

        replyButton = removeAllEventListeners(replyButton);
        onClick(replyButton, () => global.replyCollapse(id));
    };

    global.replyCollapse = (id: string) => {
        let replyButton = byId(ID_REPLY + id);
        const el = byId(ID_SUPER_CONTAINER + id);

        el.remove();
        delete shownReply[id];

        addClasses(replyButton, 'option-reply');
        removeClasses(replyButton, 'option-cancel');

        replyButton.title = 'Reply to this comment';

        replyButton = removeAllEventListeners(replyButton);
        onClick(replyButton, () => global.replyShow(id))
    }

    global.commentCollapse = (id: string) => {
        const children = byId(ID_CHILDREN + id);
        if (children) {
            addClasses(children, 'hidden');
        }

        let button = byId(ID_COLLAPSE + id);
        removeClasses(button, 'option-collapse');
        addClasses(button, 'option-uncollapse');

        button.title = 'Expand children';

        button = removeAllEventListeners(button);
        onClick(button, () => global.commentUncollapse(id));
    }

    global.commentUncollapse = (id: string) => {
        const children = byId(ID_CHILDREN + id);
        let button = byId(ID_COLLAPSE + id);

        if (children) {
            removeClasses(children, 'hidden');
        }

        removeClasses(button, 'option-uncollapse');
        addClasses(button, 'option-collapse');

        button.title = 'Collapse children';

        button = removeAllEventListeners(button);
        onClick(button, () => global.commentCollapse(id));
    }

    function parentMap(comments: Comment[]): CommentsMap {
        const m: CommentsMap = {};
        comments.forEach(comment => {
            const parentHex = comment.parentHex;
            if (!(parentHex in m)) {
                m[parentHex] = [];
            }

            comment.creationMs = new Date(comment.creationDate).getTime();

            m[parentHex].push(comment);
            commentsByHex[comment.commentHex] = {
                html: comment.html,
                markdown: comment.markdown,
            };
        });

        return m;
    }

    function commentsRender() {
        const commentsArea = byId(ID_COMMENTS_AREA);
        commentsArea.innerHTML = ''

        const cards = commentsRecurse(parentMap(comments), 'root');
        if (cards) {
            append(commentsArea, cards);
        }
    }

    function submitAuthenticated(id: string) {
        if (isAuthenticated) {
            global.commentNew(id, commenterTokenGet(), true);
            return;
        }

        global.loginBoxShow(id);
    }

    function submitAnonymous(id: string) {
        chosenAnonymous = true;
        global.commentNew(id, 'anonymous', true);
    }

    function submitAccountDecide(id: string) {
        if (requireIdentification) {
            submitAuthenticated(id);
            return;
        }

        const anonCheckbox = byId<HTMLInputElement>(ID_ANONYMOUS_CHECKBOX + id);
        const textarea = byId<HTMLTextAreaElement>(ID_TEXTAREA + id);
        const markdown = textarea.value;

        if (markdown === '') {
            addClasses(textarea, 'red-border');
            return;
        } else {
            removeClasses(textarea, 'red-border');
        }

        if (!anonCheckbox.checked) {
            submitAuthenticated(id);
        } else {
            submitAnonymous(id);
        }
    }

    // OAuth logic
    global.commentoAuth = (data: OAuthResponse): Promise<void> => {
        const provider = data.provider;
        const id = data.id;
        const popup = window.open('', '_blank');

        return apiClient.get<ApiCommenterTokenNewResponse>('commenter/token/new')
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    return;
                }
                errorHide();

                cookieSet('commentoCommenterToken', resp.commenterToken);

                popup.location = `${origin}/api/oauth/${provider}/redirect?commenterToken=${resp.commenterToken}`;

                const interval = setInterval(
                    () => {
                        if (popup.closed) {
                            clearInterval(interval);
                            selfGet()
                                .then(() => {
                                    const loggedContainer = byId(ID_LOGGED_CONTAINER);
                                    if (loggedContainer) {
                                        setAttr(loggedContainer, {style: null});
                                    }

                                    if (commenterTokenGet() !== 'anonymous') {
                                        remove(byId(ID_LOGIN));
                                    }

                                    (id ? global.commentNew(id, resp.commenterToken, false) : Promise.resolve())
                                        .then(() => global.loginBoxClose())
                                        .then(() => commentsGet())
                                        .then(() => commentsRender());
                                });
                        }
                    },
                    250);
            });
    }

    function refreshAll(): Promise<void> {
        byId(ID_ROOT).innerHTML = '';
        shownReply = {};
        return global.main();
    }

    function loginBoxCreate() {
        append(root, create('div', {id: ID_LOGIN_BOX_CONTAINER}));
    }

    global.popupRender = (id: string) => {
        const loginBoxContainer = byId(ID_LOGIN_BOX_CONTAINER);
        addClasses(loginBoxContainer, 'login-box-container');
        setAttr(loginBoxContainer, {style: 'display: none; opacity: 0;'});

        const loginBox = create('form', {id: ID_LOGIN_BOX, classes: 'login-box'});
        // This is ugly, must redesign the whole bloody login/signup form
        loginBox.addEventListener('submit', (e) => {
            e.preventDefault();
            if (!byId<HTMLButtonElement>(ID_LOGIN_BOX_PASSWORD_BUTTON)) {
                global.showPasswordField();
            } else if (popupBoxType === 'login') {
                global.login(id);
            } else {
                global.signup(id);
            }
        })

        const ssoSubtitle           = create('div',    {id: ID_LOGIN_BOX_SSO_PRETEXT, classes: 'login-box-subtitle', innerText: `Proceed with ${parent.location.host} authentication`});
        const ssoButtonContainer    = create('div',    {id: ID_LOGIN_BOX_SSO_BUTTON_CONTAINER, classes: 'oauth-buttons-container'});
        const ssoButton             = create('div',    {classes: 'oauth-buttons'});
        const hr1                   = create('hr',     {id: ID_LOGIN_BOX_HR1});
        const oauthSubtitle         = create('div',    {id: ID_LOGIN_BOX_OAUTH_PRETEXT, classes: 'login-box-subtitle', innerText: 'Proceed with social login'});
        const oauthButtonsContainer = create('div',    {id: ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER, classes: 'oauth-buttons-container'});
        const oauthButtons          = create('div',    {classes: 'oauth-buttons'});
        const hr2                   = create('hr',     {id: ID_LOGIN_BOX_HR2});
        const emailSubtitle         = create('div',    {id: ID_LOGIN_BOX_EMAIL_SUBTITLE, classes: 'login-box-subtitle', innerText: 'Login with your email address'});
        const emailButton           = create('button', {id: ID_LOGIN_BOX_EMAIL_BUTTON, type: 'submit', classes: 'email-button', innerText: 'Continue'});
        const emailContainer        = create('div', {
            classes: 'email-container',
            children: [
                create('div', {
                    classes:  'email',
                    children: [
                        create('input', {
                            id:           ID_LOGIN_BOX_EMAIL_INPUT,
                            classes:      'input',
                            name:         'email',
                            placeholder:  'Email address',
                            type:         'text',
                            autocomplete: 'email',
                        }),
                        emailButton,
                    ],
                }),
            ],
        });
        const forgotLinkContainer = create('div', {id: ID_LOGIN_BOX_FORGOT_LINK_CONTAINER, classes: 'forgot-link-container'});
        const forgotLink          = create('a',   {classes: 'forgot-link', innerText: 'Forgot your password?', parent: forgotLinkContainer});
        const loginLinkContainer  = create('div', {id: ID_LOGIN_BOX_LOGIN_LINK_CONTAINER, classes: 'login-link-container'});
        const loginLink           = create('a',   {classes: 'login-link', innerText: 'Don\'t have an account? Sign up.', parent: loginLinkContainer});
        const close               = create('div', {classes: 'login-box-close', parent: loginBox});

        addClasses(root, 'root-min-height');

        onClick(forgotLink,  () => global.forgotPassword());
        onClick(loginLink,   () => global.popupSwitch());
        onClick(close,       () => global.loginBoxClose());

        let hasOAuth = false;
        const oauthProviders = ['google', 'github', 'gitlab'];
        oauthProviders.filter(p => configuredOauths[p]).forEach(provider => {
            const button = create('button', {classes: ['button', `${provider}-button`], innerText: provider, parent: oauthButtons});
            onClick(button, () => global.commentoAuth({provider, id}));
            hasOAuth = true;
        });

        if (configuredOauths['sso']) {
            const button = create('button', {classes: ['button', 'sso-button'], innerText: 'Single Sign-On', parent: ssoButton});
            onClick(button, () => global.commentoAuth({provider: 'sso', id}));
            append(ssoButtonContainer, ssoButton);
            append(loginBox, ssoSubtitle);
            append(loginBox, ssoButtonContainer);

            if (hasOAuth || configuredOauths['commento']) {
                append(loginBox, hr1);
            }
        }

        oauthButtonsShown = hasOAuth;
        if (hasOAuth) {
            append(oauthButtonsContainer, oauthButtons);
            append(loginBox, oauthSubtitle, oauthButtonsContainer);
            if (configuredOauths['commento']) {
                append(loginBox, hr2);
            }
        }

        if (configuredOauths['commento']) {
            append(loginBox, emailSubtitle, emailContainer, forgotLinkContainer, loginLinkContainer);
        }

        popupBoxType = 'login';
        loginBoxContainer.innerHTML = '';
        append(loginBoxContainer, loginBox);
    }

    global.forgotPassword = () => {
        const popup = window.open('', '_blank');
        popup.location = `${origin}/forgot?commenter=true`;
        global.loginBoxClose();
    }

    global.popupSwitch = () => {
        const emailSubtitle = byId(ID_LOGIN_BOX_EMAIL_SUBTITLE);

        if (oauthButtonsShown) {
            remove(
                byId(ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER),
                byId(ID_LOGIN_BOX_OAUTH_PRETEXT),
                byId(ID_LOGIN_BOX_HR1),
                byId(ID_LOGIN_BOX_HR2));
        }

        if (configuredOauths['sso']) {
            remove(
                byId(ID_LOGIN_BOX_SSO_BUTTON_CONTAINER),
                byId(ID_LOGIN_BOX_SSO_PRETEXT),
                byId(ID_LOGIN_BOX_HR1),
                byId(ID_LOGIN_BOX_HR2));
        }

        remove(byId(ID_LOGIN_BOX_LOGIN_LINK_CONTAINER), byId(ID_LOGIN_BOX_FORGOT_LINK_CONTAINER));

        emailSubtitle.innerText = 'Create an account';
        popupBoxType = 'signup';
        global.showPasswordField();
        byId(ID_LOGIN_BOX_EMAIL_INPUT).focus();
    }

    function loginUP(email: string, password: string, id: string): Promise<void> {
        return apiClient.post<ApiCommenterLoginResponse>('commenter/login', {email, password})
            .then(resp => {
                if (!resp.success) {
                    global.loginBoxClose();
                    errorShow(resp.message);
                    return Promise.reject();
                }

                errorHide();
                cookieSet('commentoCommenterToken', resp.commenterToken);
                selfLoad(resp.commenter, resp.email);
                remove(byId(ID_LOGIN));
                return (id ? global.commentNew(id, resp.commenterToken, false) : undefined);
            })
            .then(() => global.loginBoxClose())
            .then(() => commentsGet())
            .then(() => commentsRender())
            .then(() => allShow());
    }

    global.login = (id: string): Promise<void> => {
        const email    = byId<HTMLInputElement>(ID_LOGIN_BOX_EMAIL_INPUT);
        const password = byId<HTMLInputElement>(ID_LOGIN_BOX_PASSWORD_INPUT);
        return loginUP(email.value, password.value, id);
    }

    global.signup = (id: string): Promise<void> => {
        const email    = byId<HTMLInputElement>(ID_LOGIN_BOX_EMAIL_INPUT);
        const name     = byId<HTMLInputElement>(ID_LOGIN_BOX_NAME_INPUT);
        const website  = byId<HTMLInputElement>(ID_LOGIN_BOX_WEBSITE_INPUT);
        const password = byId<HTMLInputElement>(ID_LOGIN_BOX_PASSWORD_INPUT);

        const data = {
            email:    email.value,
            name:     name.value,
            website:  website.value,
            password: password.value,
        };

        return apiClient.post<ApiResponseBase>('commenter/new', data)
            .then(resp => {
                if (!resp.success) {
                    global.loginBoxClose();
                    errorShow(resp.message);
                    return Promise.reject();
                }

                errorHide();
                return undefined;
            })
            .then(() => loginUP(data.email, data.password, id));
    }

    global.showPasswordField = () => {
        const isSignup = popupBoxType === 'signup';
        const loginBox = byId(ID_LOGIN_BOX);
        const subtitle = byId(ID_LOGIN_BOX_EMAIL_SUBTITLE);

        remove(
            byId(ID_LOGIN_BOX_EMAIL_BUTTON),
            byId(ID_LOGIN_BOX_LOGIN_LINK_CONTAINER),
            byId(ID_LOGIN_BOX_FORGOT_LINK_CONTAINER));
        if (oauthButtonsShown && Object.keys(configuredOauths).length) {
            remove(
                byId(ID_LOGIN_BOX_HR1),
                byId(ID_LOGIN_BOX_HR2),
                byId(ID_LOGIN_BOX_OAUTH_PRETEXT),
                byId(ID_LOGIN_BOX_OAUTH_BUTTONS_CONTAINER));
        }

        const controls = isSignup ?
            [
                {id: ID_LOGIN_BOX_NAME_INPUT,     classes: 'input', name: 'name',     type: 'text',     placeholder: 'Real Name'},
                {id: ID_LOGIN_BOX_WEBSITE_INPUT,  classes: 'input', name: 'website',  type: 'text',     placeholder: 'Website (Optional)'},
                {id: ID_LOGIN_BOX_PASSWORD_INPUT, classes: 'input', name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'new-password'},
            ] :
            [
                {id: ID_LOGIN_BOX_PASSWORD_INPUT, classes: 'input', name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'current-password'},
            ];

        subtitle.innerText = isSignup ?
            'Finish the rest of your profile to complete.' :
            'Enter your password to log in.';

        controls.forEach(c => {
            const fieldContainer = create('div', {classes: 'email-container'});
            const field          = create('div', {classes: 'email', parent: fieldContainer});
            const fieldInput     = create('input', c);
            append(field, fieldInput);
            // Add a submit button next to the password input
            if (c.type === 'password') {
                create('button', {
                    id:        ID_LOGIN_BOX_PASSWORD_BUTTON,
                    type:      'submit',
                    classes:   'email-button',
                    innerText: popupBoxType,
                    parent:    field,
                });
            }
            append(loginBox, fieldContainer);
        });

        byId(isSignup ? ID_LOGIN_BOX_NAME_INPUT : ID_LOGIN_BOX_PASSWORD_INPUT).focus();
    }

    function pageUpdate(): Promise<void> {
        const data = {
            commenterToken: commenterTokenGet(),
            domain:         parent.location.host,
            path:           pageId,
            attributes:     {isLocked, stickyCommentHex},
        };

        return apiClient.post<ApiResponseBase>('page/update', data)
            .then(resp => {
                if (!resp.success) {
                    errorShow(resp.message);
                    return Promise.reject();
                }

                errorHide();
                return undefined;
            });
    }

    global.threadLockToggle = (): Promise<void> => {
        const lock = byId<HTMLButtonElement>(ID_MOD_TOOLS_LOCK_BUTTON);
        isLocked = !isLocked;
        lock.disabled = true;
        return pageUpdate()
            .then(() => lock.disabled = false)
            .then(() => refreshAll());
    }

    global.commentSticky = (commentHex: string): Promise<void> => {
        if (stickyCommentHex !== 'none') {
            const sticky = byId(ID_STICKY + stickyCommentHex);
            removeClasses(sticky, 'option-unsticky');
            addClasses(sticky, 'option-sticky');
        }

        stickyCommentHex = stickyCommentHex === commentHex ? 'none' : commentHex;

        return pageUpdate()
            .then(() => {
                const sticky = byId(ID_STICKY + commentHex);
                if (stickyCommentHex === commentHex) {
                    removeClasses(sticky, 'option-sticky');
                    addClasses(sticky, 'option-unsticky');
                } else {
                    removeClasses(sticky, 'option-unsticky');
                    addClasses(sticky, 'option-sticky');
                }
            });
    }

    function mainAreaCreate() {
        create('div', {id: ID_MAIN_AREA, classes: 'main-area', style: 'display: none', parent: root});
    }

    function modToolsCreate() {
        const modTools = create('div', {id: ID_MOD_TOOLS, classes: 'mod-tools', style: 'display: none', parent: root});
        const lock = create('button', {id: ID_MOD_TOOLS_LOCK_BUTTON, innerHTML: isLocked ? 'Unlock Thread' : 'Lock Thread', parent: modTools});
        onClick(lock, global.threadLockToggle);
    }

    function allShow() {
        const mainArea = byId(ID_MAIN_AREA);
        const modTools = byId(ID_MOD_TOOLS);
        const loggedContainer = byId(ID_LOGGED_CONTAINER);

        setAttr(mainArea, {style: null});

        if (isModerator) {
            setAttr(modTools, {style: null});
        }

        if (loggedContainer) {
            setAttr(loggedContainer, {style: null});
        }
    }

    global.loginBoxClose = () => {
        const mainArea = byId(ID_MAIN_AREA);
        const loginBoxContainer = byId(ID_LOGIN_BOX_CONTAINER);

        removeClasses(mainArea, 'blurred');
        removeClasses(root, 'root-min-height');

        setAttr(loginBoxContainer, {style: 'display: none'});
    }

    global.loginBoxShow = (id: string) => {
        const mainArea = byId(ID_MAIN_AREA);
        const loginBoxContainer = byId(ID_LOGIN_BOX_CONTAINER);

        global.popupRender(id);

        addClasses(mainArea, 'blurred');
        setAttr(loginBoxContainer, {style: null});

        window.location.hash = ID_LOGIN_BOX_CONTAINER;

        byId(ID_LOGIN_BOX_EMAIL_INPUT).focus();
    }

    function dataTagsLoad() {
        for (const script of document.getElementsByTagName('script')) {
            if (script.src.match(/\/js\/commento\.js$/)) {
                let s = getAttr(script, 'data-page-id');
                if (s) {
                    pageId = s;
                }
                cssOverride = getAttr(script, 'data-css-override');
                autoInit = getAttr(script, 'data-auto-init') !== 'false';
                s = getAttr(script, 'data-id-root');
                if (s) {
                    ID_ROOT = s;
                }
                noFonts = getAttr(script, 'data-no-fonts') === 'true';
                hideDeleted = getAttr(script, 'data-hide-deleted') === 'true';
                break;
            }
        }
    }

    function loadHash() {
        if (window.location.hash) {
            if (window.location.hash.startsWith('#commento-')) {
                const id = window.location.hash.split('-')[1];
                const el = byId(ID_CARD + id);
                if (el === null) {
                    if (id.length === 64) {
                        // A hack to make sure it's a valid ID before showing the user a message.
                        errorShow('The comment you\'re looking for no longer exists or was deleted.');
                    }
                    return;
                }

                addClasses(el, 'highlighted-card');
                el.scrollIntoView(true);
            } else if (window.location.hash.startsWith('#commento')) {
                root.scrollIntoView(true);
            }
        }
    }

    global.main = (): Promise<void> => {
        root = byId(ID_ROOT);
        if (!root) {
            console.error(`[commento] No root element with ID '${ID_ROOT}' found`);
            return Promise.reject();
        }

        if (mobileView === null) {
            mobileView = root.getBoundingClientRect()['width'] < 450;
        }

        addClasses(root, ['root', !noFonts && 'root-font']);

        loginBoxCreate();
        errorElementCreate();
        mainAreaCreate();

        // Begin by loading the stylesheet
        return cssLoad(`${cdn}/css/commento.css`)
            // Load stylesheet override, if any
            .then(() => cssOverride && cssLoad(cssOverride))
            // Load information about ourselves
            .then(() => selfGet())
            // Fetch comments
            .then(() => commentsGet())
            // Create the layout
            .then(() => {
                modToolsCreate();
                rootCreate();
                commentsRender();
                append(root, footerLoad());
                loadHash();
                allShow();
                nameWidthFix();
            });
    }

    let initted = false;

    function init() {
        if (initted) {
            return;
        }
        initted = true;

        dataTagsLoad();

        if (autoInit) {
            global.main();
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

}((window as any).commento, document));
