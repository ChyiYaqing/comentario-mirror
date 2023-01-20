import { HttpClient } from './http-client';
import { Comment, Commenter, CommentMap, CommentsMap, Email, SortPolicy, SortPolicyProps } from './models';
import {
    ApiCommentEditResponse,
    ApiCommenterLoginResponse,
    ApiCommenterTokenNewResponse,
    ApiCommentListResponse,
    ApiCommentNewResponse,
    ApiResponseBase,
    ApiSelfResponse,
} from './api';

const IDS = {
    mainArea:                      'main-area',
    login:                         'login',
    loginBoxContainer:             'login-box-container',
    loginBox:                      'login-box',
    loginBoxEmailSubtitle:         'login-box-email-subtitle',
    loginBoxEmailInput:            'login-box-email-input',
    loginBoxPasswordButton:        'login-box-password-button',
    loginBoxPasswordInput:         'login-box-password-input',
    loginBoxNameInput:             'login-box-name-input',
    loginBoxWebsiteInput:          'login-box-website-input',
    loginBoxEmailButton:           'login-box-email-button',
    loginBoxForgotLinkContainer:   'login-box-forgot-link-container',
    loginBoxLoginLinkContainer:    'login-box-login-link-container',
    loginBoxSsoPretext:            'login-box-sso-pretext',
    loginBoxSsoButtonContainer:    'login-box-sso-button-container',
    loginBoxHr1:                   'login-box-hr1',
    loginBoxOauthPretext:          'login-box-oauth-pretext',
    loginBoxOauthButtonsContainer: 'login-box-oauth-buttons-container',
    loginBoxHr2:                   'login-box-hr2',
    modTools:                      'mod-tools',
    modToolsLockButton:            'mod-tools-lock-button',
    error:                         'error',
    loggedContainer:               'logged-container',
    preCommentsArea:               'pre-comments-area',
    commentsArea:                  'comments-area',
    superContainer:                'textarea-super-container-',
    textareaContainer:             'textarea-container-',
    textarea:                      'textarea-',
    anonymousCheckbox:             'anonymous-checkbox-',
    sortPolicy:                    'sort-policy-',
    card:                          'comment-card-',
    body:                          'comment-body-',
    text:                          'comment-text-',
    subtitle:                      'comment-subtitle-',
    timeago:                       'comment-timeago-',
    score:                         'comment-score-',
    options:                       'comment-options-',
    edit:                          'comment-edit-',
    reply:                         'comment-reply-',
    collapse:                      'comment-collapse-',
    upvote:                        'comment-upvote-',
    downvote:                      'comment-downvote-',
    approve:                       'comment-approve-',
    remove:                        'comment-remove-',
    sticky:                        'comment-sticky-',
    children:                      'comment-children-',
    contents:                      'comment-contents-',
    name:                          'comment-name-',
    submitButton:                  'submit-button-',
    markdownButton:                'markdown-button-',
    markdownHelp:                  'markdown-help-',
    footer:                        'footer',
};

type CallbackFunction = () => void;

interface ElementConfig {
    /** ID to assign to the created element (excluding the prefix). */
    id?:        string;
    classes?:   string | string[];
    innerText?: string;
    innerHTML?: string;
    parent?:    HTMLElement;
    children?:  HTMLElement[];

    [ k: string ]: any;
}

export class Comentario {

    private static readonly idPrefix = 'comentario-';

    private readonly origin = '[[[.Origin]]]';
    private readonly cdn = '[[[.CdnPrefix]]]';

    /** HTTP client we'll use for API requests. */
    private readonly apiClient = new HttpClient(`${this.origin}/api`);

    /** Default ID of the container element Comentario will be embedded into. */
    private rootId = 'commento';

    private root: HTMLElement = null;
    private pageId = parent.location.pathname;
    private cssOverride: string;
    private noFonts = false;
    private hideDeleted = false;
    private autoInit = true;
    private isAuthenticated = false;
    private comments: Comment[] = [];
    private readonly commentsByHex: CommentMap = {};
    private readonly commenters: { [k: string]: Commenter } = {};
    private requireIdentification = true;
    private isModerator = false;
    private isFrozen = false;
    private chosenAnonymous = false;
    private isLocked = false;
    private stickyCommentHex = 'none';
    private shownReply: { [k: string]: boolean };
    private readonly shownEdit: { [k: string]: boolean } = {};
    private configuredOauths: { [k: string]: boolean } = {};
    private anonymousOnly = false;
    private popupBoxType = 'login';
    private oauthButtonsShown = false;
    private sortPolicy: SortPolicy = 'score-desc';
    private selfHex: string = undefined;
    private mobileView: boolean | null = null;
    private readonly loadedCss: { [k: string]: boolean } = {};
    private initialised = false;

    private readonly sortingProps: { [k in SortPolicy]: SortPolicyProps<Comment> } = {
        'score-desc':        {label: 'Upvotes', comparator: (a, b) => b.score - a.score},
        'creationdate-desc': {label: 'Newest',  comparator: (a, b) => a.creationMs < b.creationMs ? 1 : -1},
        'creationdate-asc':  {label: 'Oldest',  comparator: (a, b) => a.creationMs < b.creationMs ? -1 : 1},
    };

    constructor(
        private readonly doc: Document,
    ) {
        this.whenDocReady().then(() => this.init());
    }

    /**
     * The main worker routine of Comentario
     * @return Promise that resolves as soon as Comentario setup is complete
     */
    main(): Promise<void> {
        this.root = this.byId(this.rootId, true);
        if (!this.root) {
            return this.reject(`No root element with id='${this.rootId}' found. Check your configuration and HTML.`);
        }

        if (this.mobileView === null) {
            this.mobileView = this.root.getBoundingClientRect()['width'] < 450;
        }

        this.addClasses(this.root, ['root', !this.noFonts && 'root-font']);

        // Begin by loading the stylesheet
        return this.cssLoad(`${this.cdn}/css/commento.css`)
            // Load stylesheet override, if any
            .then(() => this.cssOverride && this.cssLoad(this.cssOverride))
            // Load the UI
            .then(() => this.reload());
    }

    /**
     * Reload the app UI.
     */
    private reload() {
        // Remove any content from the root
        this.byId(this.rootId, true).innerHTML = '';
        this.shownReply = {};

        // Create base elements
        this.loginBoxCreate();
        this.errorElementCreate();
        this.mainAreaCreate();

        // Load information about ourselves
        return this.selfGet()
            // Fetch comments
            .then(() => this.commentsGet())
            // Create the layout
            .then(() => {
                this.modToolsCreate();
                this.rootCreate();
                this.commentsRender();
                this.append(this.root, this.footerLoad());
                this.loadHash();
                this.allShow();
                this.nameWidthFix();
            });
    }

    /**
     * Return a rejected promise with the given message.
     * @param message Message to reject the promise with.
     * @private
     */
    private reject(message: string): Promise<never> {
        return Promise.reject(`Comentario: ${message}`);
    }

    /**
     * Returns a promise that gets resolved as soon as the document reaches at least its 'interactive' state.
     * @private
     */
    private whenDocReady(): Promise<void> {
        return new Promise(resolved => {
            const checkState = () => {
                switch (this.doc.readyState) {
                    // The document is still loading. The div we need to fill might not have been parsed yet, so let's
                    // wait and retry when the readyState changes
                    case 'loading':
                        this.doc.addEventListener('readystatechange', () => checkState());
                        break;

                    case 'interactive': // The document has been parsed and DOM objects are now accessible.
                    case 'complete': // The page has fully loaded (including JS, CSS, and images)
                        resolved();
                }
            };
            checkState();
        });
    }

    private init(): Promise<void> {
        // Only perform initialisation once
        if (this.initialised) {
            return this.reject('Already initialised, ignoring the repeated init call');
        }

        this.initialised = true;

        // Parse any custom data-* tags on the Comentario script element
        this.dataTagsLoad();

        // If automatic initialisation is activated (default), run Comentario
        return this.autoInit ? this.main() : Promise.resolve();
    }

    /**
     * Finds and returns an HTML element with the given ID (optionally prepending it with idPrefix), or null if no such
     * element exists.
     * @param id ID of the element to find (excluding the prefix).
     * @param noPrefix Whether skip prepending the ID with idPrefix.
     */
    byId<T extends HTMLElement>(id: string, noPrefix?: boolean): T {
        return this.doc.getElementById(noPrefix ? id : Comentario.idPrefix + id) as T;
    }

    prepend(root: HTMLElement, el: HTMLElement) {
        root.prepend(el);
    }

    append(parent: HTMLElement, ...children: HTMLElement[]) {
        children.forEach(c => parent.appendChild(c));
    }

    insertAfter(el1: HTMLElement, el2: HTMLElement) {
        el1.parentNode.insertBefore(el2, el1.nextSibling);
    }

    /**
     * Add the provided class or classes to the element.
     * @param el Element to add classes to.
     * @param classes string|array Class(es) to add. Falsy values are ignored.
     */
    addClasses(el: HTMLElement, classes: string | string[]) {
        (Array.isArray(classes) ? classes : [classes]).forEach(c => c && el.classList.add(`commento-${c}`));
    }

    /**
     * Remove the provided class or classes from the element.
     * @param el Element to remove classes from.
     * @param classes string|array Class(es) to remove. Falsy values are ignored.
     */
    removeClasses(el: HTMLElement, classes: string | string[]) {
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
    createElement<K extends keyof HTMLElementTagNameMap>(tagName: K, config?: ElementConfig): HTMLElementTagNameMap[K] {
        // Create a new HTML element
        const e = this.doc.createElement(tagName) as HTMLElementTagNameMap[K];

        // If there's any config passed
        if (config) {
            // Set up the ID, if given, and clean it up from the config
            if ('id' in config) {
                e.id = Comentario.idPrefix + config.id;
                delete config.id;
            }

            // Set up the classes, if given, and clean them up from the config
            if ('classes' in config) {
                this.addClasses(e, config.classes);
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
            this.setAttr(e, config);

            // Add the child to the parent, if any
            if (parent) {
                parent.appendChild(e);
            }
        }
        return e;
    }

    remove(...elements: HTMLElement[]) {
        elements?.forEach(e => e && e.parentNode.removeChild(e));
    }

    getAttr(node: HTMLElement, attrName: string) {
        const attr = node.attributes.getNamedItem(attrName);
        return attr === undefined ? undefined : attr?.value;
    }

    removeAllEventListeners<T extends HTMLElement>(node: T): T {
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
     * Bind a handler to the this.onClick event of the given element.
     * @param e Element to bind a handler to.
     * @param handler Handler to bind.
     */
    onClick(e: HTMLElement, handler: CallbackFunction) {
        e.addEventListener('click', handler, false);
    }

    /**
     * Set node attributes from the provided object.
     * @param node HTML element to set attributes on.
     * @param values Object that provides attribute names (keys) and their values. null and undefined values cause attribute removal from the node.
     */
    setAttr(node: HTMLElement, values: { [k: string]: string }) {
        if (node) {
            Object.keys(values).forEach(k => {
                const v = values[k];
                if (v === undefined || v === null) {
                    node.removeAttribute(k);
                } else {
                    node.setAttribute(k, v);
                }
            });
        }
    }

    cookieGet(name: string): string {
        const c = `; ${this.doc.cookie}`;
        const x = c.split(`; ${name}=`);
        return x.length === 2 ? x.pop().split(';').shift() : null;
    }

    cookieSet(name: string, value: string) {
        const date = new Date();
        date.setTime(date.getTime() + (365 * 24 * 60 * 60 * 1000));
        this.doc.cookie = `${name}=${value}; expires=${date.toUTCString()}; path=/`;
    }

    commenterTokenGet() {
        const commenterToken = this.cookieGet('commentoCommenterToken');
        return commenterToken === undefined ? 'anonymous' : commenterToken;
    }

    logout(): Promise<void> {
        this.cookieSet('commentoCommenterToken', 'anonymous');
        this.isAuthenticated = false;
        this.isModerator = false;
        this.selfHex = undefined;
        return this.reload();
    }

    profileEdit() {
        window.open(`${this.origin}/profile?commenterToken=${this.commenterTokenGet()}`, '_blank');
    }

    notificationSettings(unsubscribeSecretHex: string) {
        window.open(`${this.origin}/unsubscribe?unsubscribeSecretHex=${unsubscribeSecretHex}`, '_blank');
    }

    selfLoad(commenter: Commenter, email: Email) {
        this.commenters[commenter.commenterHex] = commenter;
        this.selfHex = commenter.commenterHex;

        const loggedContainer = this.createElement('div', {id: IDS.loggedContainer, classes: 'logged-container', style: 'display: none'});
        const loggedInAs      = this.createElement('div', {classes: 'logged-in-as', parent: loggedContainer});
        const name            = this.createElement(commenter.link !== 'undefined' ? 'a' : 'div', {classes: 'name', innerText: commenter.name, parent: loggedInAs});
        const btnSettings     = this.createElement('div', {classes: 'profile-button', innerText: 'Notification Settings'});
        const btnEditProfile  = this.createElement('div', {classes: 'profile-button', innerText: 'Edit Profile'});
        const btnLogout       = this.createElement('div', {classes: 'profile-button', innerText: 'Logout', parent: loggedContainer});
        const color = this.colorGet(`${commenter.commenterHex}-${commenter.name}`);

        // Set the profile href for the commenter, if any
        if (commenter.link !== 'undefined') {
            this.setAttr(name, {href: commenter.link});
        }

        this.onClick(btnLogout,      () => this.logout());
        this.onClick(btnSettings,    () => this.notificationSettings(email.unsubscribeSecretHex));
        this.onClick(btnEditProfile, () => this.profileEdit);

        // Add an avatar
        if (commenter.photo === 'undefined') {
            this.createElement('div', {
                classes:   'avatar',
                innerHTML: commenter.name[0].toUpperCase(),
                style:     `background-color: ${color}`,
                parent:    loggedInAs,
            });
        } else {
            this.createElement('img', {
                classes: 'avatar-img',
                src:     `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                loading: 'lazy',
                alt:     '',
                parent:  loggedInAs,
            });
        }

        // If it's a local user, add an Edit profile button
        if (commenter.provider === 'commento') {
            this.append(loggedContainer, btnEditProfile);
        }
        this.append(loggedContainer, btnSettings);

        // Add the container to the root
        this.prepend(this.root, loggedContainer);
        this.isAuthenticated = true;
    }

    selfGet(): Promise<void> {
        const commenterToken = this.commenterTokenGet();
        if (commenterToken === 'anonymous') {
            this.isAuthenticated = false;
            return Promise.resolve();
        }

        return this.apiClient.post<ApiSelfResponse>('commenter/self', {commenterToken: this.commenterTokenGet()})
            .then(resp => {
                if (!resp.success) {
                    this.cookieSet('commentoCommenterToken', 'anonymous');
                    return this.reject(resp.message);
                }

                this.selfLoad(resp.commenter, resp.email);
                this.allShow();
                return undefined;
            });
    }

    /**
     * Load the stylesheet with the provided URL into the DOM
     * @param url Stylesheet URL.
     */
    cssLoad(url: string): Promise<void> {
        // Don't bother if the stylesheet has been loaded already
        return this.loadedCss[url] ?
            Promise.resolve() :
            new Promise(resolve => {
                this.loadedCss[url] = true;
                const link = this.createElement('link', {href: url, rel: 'stylesheet', type: 'text/css'});
                link.addEventListener('load', () => resolve());
                this.append(this.doc.getElementsByTagName('head')[0], link);
            });
    }

    footerLoad() {
        return this.createElement('div', {
            id:       IDS.footer,
            classes:  'footer',
            children: [
                this.createElement('div', {
                    classes:  'logo-container',
                    children: [
                        this.createElement('a', {
                            classes:  'logo',
                            href:     'https://comentario.app/',
                            target:   '_blank',
                            children: [
                                this.createElement('span', {classes: 'logo-text', innerText: 'Comentario 🗨'}),
                            ],
                        }),
                    ],
                }),
            ],
        });
    }

    commentsGet(): Promise<void> {
        return this.apiClient.post<ApiCommentListResponse>(
            'comment/list',
            {
                commenterToken: this.commenterTokenGet(),
                domain:         parent.location.host,
                path:           this.pageId,
            })
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return;
                }

                this.errorHide();

                Object.assign(this.commenters, resp.commenters);
                this.requireIdentification = resp.requireIdentification;
                this.isModerator = resp.isModerator;
                this.isFrozen = resp.isFrozen;
                this.isLocked = resp.attributes.isLocked;
                this.stickyCommentHex = resp.attributes.stickyCommentHex;
                this.comments = resp.comments;
                this.configuredOauths = resp.configuredOauths;
                this.sortPolicy = resp.defaultSortPolicy;
            });
    }

    errorShow(text: string) {
        const el = this.byId<HTMLDivElement>(IDS.error);
        el.innerText = text;
        this.setAttr(el, {style: 'display: block;'});
    }

    errorHide() {
        this.setAttr(this.byId(IDS.error), {style: 'display: none;'});
    }

    errorElementCreate() {
        this.createElement('div', {id: IDS.error, classes: 'error-box', style: 'display: none;', parent: this.root});
    }

    autoExpander(el: HTMLElement): CallbackFunction {
        return () => {
            el.style.height = '';
            el.style.height = `${Math.min(Math.max(el.scrollHeight, 75), 400)}px`;
        };
    }

    markdownHelpShow(id: string) {
        this.createElement('table', {
            id:       IDS.markdownHelp + id,
            classes:  'markdown-help',
            parent:   this.byId(IDS.superContainer + id),
            children: [
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<i>italics</i>'}),
                        this.createElement('td', {innerHTML: 'surround text with <pre>*asterisks*</pre>'}),
                    ],
                }),
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<b>bold</b>'}),
                        this.createElement('td', {innerHTML: 'surround text with <pre>**two asterisks**</pre>'}),
                    ],
                }),
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<pre>code</pre>'}),
                        this.createElement('td', {innerHTML: 'surround text with <pre>`backticks`</pre>'}),
                    ],
                }),
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<del>strikethrough</del>'}),
                        this.createElement('td', {innerHTML: 'surround text with <pre>~~two tilde characters~~</pre>'}),
                    ],
                }),
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<a href="https://example.com">hyperlink</a>'}),
                        this.createElement('td', {innerHTML: '<pre>[hyperlink](https://example.com)</pre> or just a bare URL'}),
                    ],
                }),
                this.createElement('tr', {
                    children: [
                        this.createElement('td', {innerHTML: '<blockquote>quote</blockquote>'}),
                        this.createElement('td', {innerHTML: 'prefix with <pre>&gt;</pre>'}),
                    ],
                }),
            ],
        });

        // Add a collapse button
        const markdownButton = this.removeAllEventListeners(this.byId<HTMLAnchorElement>(IDS.markdownButton + id));
        this.onClick(markdownButton, () => this.markdownHelpHide(id));
    }

    markdownHelpHide(id: string) {
        let markdownButton = this.byId<HTMLAnchorElement>(IDS.markdownButton + id);
        const markdownHelp = this.byId(IDS.markdownHelp + id);

        markdownButton = this.removeAllEventListeners(markdownButton);
        this.onClick(markdownButton, () => this.markdownHelpShow(id));
        this.remove(markdownHelp);
    }

    /**
     * Create a new editor for editing comment text.
     * @param commentHex Comment's hex ID.
     * @param isEdit Whether it's adding a new comment (false) or editing an existing one (true)
     */
    textareaCreate(commentHex: string, isEdit: boolean): HTMLDivElement {
        const textOuter        = this.createElement('div',      {id: IDS.superContainer + commentHex, classes: 'button-margin'});
        const textCont         = this.createElement('div',      {id: IDS.textareaContainer + commentHex, classes: 'textarea-container', parent: textOuter});
        const textArea         = this.createElement('textarea', {id: IDS.textarea + commentHex, placeholder: 'Add a comment', parent: textCont});
        const anonCheckbox     = this.createElement('input',    {id: IDS.anonymousCheckbox + commentHex, type: 'checkbox'});
        const anonCheckboxCont = this.createElement('div', {
            classes:  ['round-check', 'anonymous-checkbox-container'],
            children: [
                anonCheckbox,
                this.createElement(
                    'label',
                    {for: Comentario.idPrefix + IDS.anonymousCheckbox + commentHex, innerText: 'Comment anonymously'}),
            ],
        });
        const submitButton = this.createElement('button', {
            id:        IDS.submitButton + commentHex,
            type:      'submit',
            classes:   ['button', 'submit-button'],
            innerText: isEdit ? 'Save Changes' : 'Add Comment',
            parent:    textOuter,
        });
        const markdownButton = this.createElement('a', {
            id:        IDS.markdownButton + commentHex,
            classes:   'markdown-button',
            innerHTML: '<b>M⬇</b>&nbsp;Markdown',
        });

        if (this.anonymousOnly) {
            anonCheckbox.checked = true;
            anonCheckbox.setAttribute('disabled', 'true');
        }

        textArea.oninput = this.autoExpander(textArea);
        this.onClick(submitButton, () => isEdit ? this.saveCommentEdits(commentHex) : this.submitAccountDecide(commentHex));
        this.onClick(markdownButton, () => this.markdownHelpShow(commentHex));
        if (!this.requireIdentification && !isEdit) {
            this.append(textOuter, anonCheckboxCont);
        }
        this.append(textOuter, markdownButton);
        return textOuter;
    }

    sortPolicyApply(policy: SortPolicy) {
        this.removeClasses(this.byId(IDS.sortPolicy + this.sortPolicy), 'sort-policy-button-selected');

        const commentsArea = this.byId<HTMLDivElement>(IDS.commentsArea);
        commentsArea.innerHTML = '';
        this.sortPolicy = policy;
        const cards = this.commentsRecurse(this.parentMap(this.comments), 'root');
        if (cards) {
            this.append(commentsArea, cards);
        }

        this.addClasses(this.byId(IDS.sortPolicy + policy), 'sort-policy-button-selected');
    }

    sortPolicyBox(): HTMLDivElement {
        const container = this.createElement('div', {classes: 'sort-policy-buttons-container'});
        const buttonBar = this.createElement('div', {classes: 'sort-policy-buttons', parent: container});
        Object.keys(this.sortingProps).forEach((sp: SortPolicy) => {
            const sortPolicyButton = this.createElement('a', {
                id:        IDS.sortPolicy + sp,
                classes:   ['sort-policy-button', sp === this.sortPolicy && 'sort-policy-button-selected'],
                innerText: this.sortingProps[sp].label,
                parent:    buttonBar,
            });
            this.onClick(sortPolicyButton, () => this.sortPolicyApply(sp));
        });
        return container;
    }

    /**
     * Create the top-level ("main area") elements in the root.
     */
    rootCreate(): void {
        const mainArea = this.byId(IDS.mainArea);
        const login           = this.createElement('div', {id: IDS.login, classes: 'login'});
        const loginText       = this.createElement('div', {classes: 'login-text', innerText: 'Login'});
        const preCommentsArea = this.createElement('div', {id: IDS.preCommentsArea});
        const commentsArea    = this.createElement('div', {id: IDS.commentsArea, classes: 'comments'});
        this.onClick(loginText, () => this.loginBoxShow(null));

        // If there's an OAuth provider configured, add a Login button
        if (Object.keys(this.configuredOauths).some(k => this.configuredOauths[k])) {
            this.append(login, loginText);
        } else if (!this.requireIdentification) {
            this.anonymousOnly = true;
        }

        if (this.isLocked || this.isFrozen) {
            if (this.isAuthenticated || this.chosenAnonymous) {
                this.append(mainArea, this.messageCreate('This thread is locked. You cannot add new comments.'));
                this.remove(this.byId(IDS.login));
            } else {
                // Add a root editor (for creating a new comment)
                this.append(mainArea, login, this.textareaCreate('root', false));
            }
        } else {
            if (this.isAuthenticated) {
                this.remove(this.byId(IDS.login));
            } else {
                this.append(mainArea, login);
            }
            // Add a root editor (for creating a new comment)
            this.append(mainArea, this.textareaCreate('root', false));
        }

        // If there's any comment, add sort buttons
        if (this.comments.length > 0) {
            this.append(mainArea, this.sortPolicyBox());
        }
        this.append(mainArea, preCommentsArea, commentsArea);
        this.append(this.root, mainArea);
    }

    messageCreate(text: string): HTMLDivElement {
        return this.createElement('div', {classes: 'moderation-notice', innerText: text});
    }

    commentNew(commentHex: string, commenterToken: string, appendCard: boolean): Promise<void> {
        const container   = this.byId<HTMLDivElement>(IDS.superContainer + commentHex);
        const textarea    = this.byId<HTMLTextAreaElement>(IDS.textarea + commentHex);
        const replyButton = this.byId<HTMLButtonElement>(IDS.reply + commentHex);

        const markdown = textarea.value;

        if (markdown === '') {
            this.addClasses(textarea, 'red-border');
            return Promise.reject();
        }

        this.removeClasses(textarea, 'red-border');

        const data = {
            commenterToken,
            domain: parent.location.host,
            path: this.pageId,
            parentHex: commentHex,
            markdown,
        };

        return this.apiClient.post<ApiCommentNewResponse>('comment/new', data)
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return;
                }

                this.errorHide();

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
                    this.prepend(this.byId(IDS.superContainer + commentHex), this.messageCreate(message));
                }

                const comment: Comment = {
                    commentHex: resp.commentHex,
                    commenterHex: this.selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : this.selfHex,
                    markdown,
                    html: resp.html,
                    parentHex: 'root',
                    score: 0,
                    state: 'approved',
                    direction: 0,
                    creationDate: new Date().toISOString(),
                };

                const newCard = this.commentsRecurse({root: [comment]}, 'root');

                this.commentsByHex[resp.commentHex] = comment;
                if (appendCard) {
                    if (commentHex !== 'root') {
                        container.replaceWith(newCard);

                        this.shownReply[commentHex] = false;

                        this.addClasses(replyButton, 'option-reply');
                        this.removeClasses(replyButton, 'option-cancel');

                        replyButton.title = 'Reply to this comment';

                        this.onClick(replyButton, () => this.replyShow(commentHex));
                    } else {
                        textarea.value = '';
                        this.insertAfter(this.byId(IDS.preCommentsArea), newCard);
                    }
                } else if (commentHex === 'root') {
                    textarea.value = '';
                }
            });
    }

    colorGet(name: string) {
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

    timeDifference(current: number, previous: number): string {
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

    scorify(score: number) {
        return score === 1 ? 'One point' : `${score} points`;
    }

    commentsRecurse(parentMap: CommentsMap, parentHex: string) {
        const cur = parentMap[parentHex];
        if (!cur || !cur.length) {
            return null;
        }

        cur.sort((a, b) => {
            return !a.deleted && a.commentHex === this.stickyCommentHex ?
                -Infinity :
                !b.deleted && b.commentHex === this.stickyCommentHex ?
                    Infinity :
                    this.sortingProps[this.sortPolicy].comparator(a, b);
        });

        const curTime = (new Date()).getTime();
        const cards = this.createElement('div');
        cur.forEach(comment => {
            const commenter = this.commenters[comment.commenterHex];
            const hex = comment.commentHex;
            const header = this.createElement('div', {classes: 'header'});
            const name = this.createElement(
                commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '' ? 'a' : 'div',
                {
                    id:        IDS.name + hex,
                    innerText: comment.deleted ? '[deleted]' : commenter.name,
                    classes:   'name',
                });
            const color = this.colorGet(`${comment.commenterHex}-${commenter.name}`);
            const card     = this.createElement('div', {id: IDS.card     + hex, style: `border-left: 2px solid ${color}`, classes: 'card'});
            const subtitle = this.createElement('div', {id: IDS.subtitle + hex, classes: 'subtitle'});
            const timeago = this.createElement('div', {
                id:        IDS.timeago + hex,
                classes:   'timeago',
                innerHTML: this.timeDifference(curTime, comment.creationMs),
                title:     comment.creationDate.toString(),
            });
            const score = this.createElement('div', {id: IDS.score + hex, classes: 'score', innerText: this.scorify(comment.score)});
            const body     = this.createElement('div',    {id: IDS.body     + hex, classes: 'body'});
            const text     = this.createElement('div',    {id: IDS.text     + hex, innerHTML: comment.html});
            const options  = this.createElement('div',    {id: IDS.options  + hex, classes: 'options'});
            const edit     = this.createElement('button', {id: IDS.edit     + hex, type: 'button', classes: ['option-button', 'option-edit'],     title: 'Edit'});
            const reply    = this.createElement('button', {id: IDS.reply    + hex, type: 'button', classes: ['option-button', 'option-reply'],    title: 'Reply'});
            const collapse = this.createElement('button', {id: IDS.collapse + hex, type: 'button', classes: ['option-button', 'option-collapse'], title: 'Collapse children'});
            let   upvote   = this.createElement('button', {id: IDS.upvote   + hex, type: 'button', classes: ['option-button', 'option-upvote'],   title: 'Upvote'});
            let   downvote = this.createElement('button', {id: IDS.downvote + hex, type: 'button', classes: ['option-button', 'option-downvote'], title: 'Downvote'});
            const approve  = this.createElement('button', {id: IDS.approve  + hex, type: 'button', classes: ['option-button', 'option-approve'],  title: 'Approve'});
            const remove   = this.createElement('button', {id: IDS.remove   + hex, type: 'button', classes: ['option-button', 'option-remove'],   title: 'Remove'});
            const sticky   = this.createElement('button', {
                id:      IDS.sticky + hex,
                type:    'button',
                classes: ['option-button', this.stickyCommentHex === hex ? 'option-unsticky' : 'option-sticky'],
                title:   this.stickyCommentHex === hex ? this.isModerator ? 'Unsticky' : 'This comment has been stickied' : 'Sticky',
            });
            const contents = this.createElement('div',    {id: IDS.contents + hex});
            if (this.mobileView) {
                this.addClasses(options, 'options-mobile');
            }

            const children = this.commentsRecurse(parentMap, hex);
            if (children) {
                children.id = IDS.children + hex;
            }

            let avatar;
            if (commenter.photo === 'undefined') {
                avatar = this.createElement('div', {style: `background-color: ${color}`, classes: 'avatar'});

                if (comment.commenterHex === 'anonymous') {
                    avatar.innerHTML = '?';
                    avatar.style.fontWeight = 'bold';
                } else {
                    avatar.innerHTML = commenter.name[0].toUpperCase();
                }
            } else {
                this.createElement('img', {
                    src:     `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                    classes: 'avatar-img',
                });
            }
            if (this.isModerator && comment.state !== 'approved') {
                this.addClasses(card, 'dark-card');
            }
            if (commenter.isModerator) {
                this.addClasses(name, 'moderator');
            }
            if (comment.state === 'flagged') {
                this.addClasses(name, 'flagged');
            }

            if (this.isAuthenticated) {
                if (comment.direction > 0) {
                    this.addClasses(upvote, 'upvoted');
                } else if (comment.direction < 0) {
                    this.addClasses(downvote, 'downvoted');
                }
            }

            // Add comment toolbar buttons
            this.onClick(edit,     () => this.startEditing(hex));
            this.onClick(collapse, () => this.commentCollapse(hex));
            this.onClick(approve,  () => this.commentApprove(hex));
            this.onClick(remove,   () => this.commentDelete(hex));
            this.onClick(sticky,   () => this.commentSticky(hex));

            if (this.isAuthenticated) {
                const upDown = this.upDownOnClickSet(upvote, downvote, hex, comment.direction);
                upvote = upDown[0];
                downvote = upDown[1];
            } else {
                this.onClick(upvote,   () => this.loginBoxShow(null));
                this.onClick(downvote, () => this.loginBoxShow(null));
            }

            this.onClick(reply, () => this.replyShow(hex));

            if (commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '') {
                this.setAttr(name, {href: commenter.link});
            }

            this.append(options, collapse);

            if (!comment.deleted) {
                this.append(options, downvote, upvote);
            }

            if (comment.commenterHex === this.selfHex) {
                this.append(options, edit);
            } else if (!comment.deleted) {
                this.append(options, reply);
            }

            if (!comment.deleted && (this.isModerator && parentHex === 'root')) {
                this.append(options, sticky);
            }

            if (!comment.deleted && (this.isModerator || comment.commenterHex === this.selfHex)) {
                this.append(options, remove);
            }

            if (this.isModerator && comment.state !== 'approved') {
                this.append(options, approve);
            }

            if (!comment.deleted && (!this.isModerator && this.stickyCommentHex === hex)) {
                this.append(options, sticky);
            }

            this.setAttr(options, {style: `width: ${(options.childNodes.length + 1) * 32}px;`});
            for (let i = 0; i < options.childNodes.length; i++) {
                this.setAttr(options.children[i] as HTMLElement, {style: `right: ${i * 32}px;`});
            }

            this.append(subtitle, score, timeago);

            if (!this.mobileView) {
                this.append(header, options);
            }
            this.append(header, avatar, name, subtitle);
            this.append(body, text);
            this.append(contents, body);
            if (this.mobileView) {
                this.append(contents, options);
                this.createElement('div', {classes: 'options-clearfix', parent: contents});
            }

            if (children) {
                this.addClasses(children, 'body');
                this.append(contents, children);
            }

            this.append(card, header, contents);

            if (comment.deleted && (this.hideDeleted || children === null)) {
                return;
            }

            this.append(cards, card);
        });

        return cards.childNodes.length ? cards : null;
    }

    commentApprove(commentHex: string): Promise<void> {
        return this.apiClient.post<ApiResponseBase>(
            'comment/approve',
            {commenterToken: this.commenterTokenGet(), commentHex},
        )
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return;
                }
                this.errorHide();

                const card = this.byId(IDS.card + commentHex);
                const name = this.byId(IDS.name + commentHex);
                const tick = this.byId(IDS.approve + commentHex);

                this.removeClasses(card, 'dark-card');
                this.removeClasses(name, 'flagged');
                this.remove(tick);
            });
    }

    commentDelete(commentHex: string): Promise<void> {
        if (!confirm('Are you sure you want to delete this comment?')) {
            return Promise.reject();
        }

        return this.apiClient.post<ApiResponseBase>('comment/delete', {commenterToken: this.commenterTokenGet(), commentHex})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return;
                }

                this.errorHide();
                const text = this.byId<HTMLDivElement>(IDS.text + commentHex);
                text.innerText = '[deleted]';
            });
    }

    nameWidthFix() {
        const els = this.doc.getElementsByClassName('commento-name');

        for (let i = 0; i < els.length; i++) {
            this.setAttr(els[i] as HTMLElement, {style: `max-width: ${els[i].getBoundingClientRect()['width'] + 20}px;`});
        }
    }

    upDownOnClickSet(upvote: HTMLButtonElement, downvote: HTMLButtonElement, commentHex: string, direction: number): [HTMLButtonElement, HTMLButtonElement] {
        upvote = this.removeAllEventListeners(upvote);
        downvote = this.removeAllEventListeners(downvote);

        if (direction > 0) {
            this.onClick(upvote,   () => this.vote(commentHex, 1, 0));
            this.onClick(downvote, () => this.vote(commentHex, 1, -1));
        } else if (direction < 0) {
            this.onClick(upvote,   () => this.vote(commentHex, -1, 1));
            this.onClick(downvote, () => this.vote(commentHex, -1, 0));
        } else {
            this.onClick(upvote,   () => this.vote(commentHex, 0, 1));
            this.onClick(downvote, () => this.vote(commentHex, 0, -1));
        }

        return [upvote, downvote];
    }

    vote(commentHex: string, oldDirection: number, direction: number): Promise<void> {
        let upvote   = this.byId<HTMLButtonElement>(IDS.upvote + commentHex);
        let downvote = this.byId<HTMLButtonElement>(IDS.downvote + commentHex);
        const score  = this.byId<HTMLDivElement>(IDS.score + commentHex);

        const upDown = this.upDownOnClickSet(upvote, downvote, commentHex, direction);
        upvote = upDown[0];
        downvote = upDown[1];

        this.removeClasses(upvote, 'upvoted');
        this.removeClasses(downvote, 'downvoted');
        if (direction > 0) {
            this.addClasses(upvote, 'upvoted');
        } else if (direction < 0) {
            this.addClasses(downvote, 'downvoted');
        }

        score.innerText = this.scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) + direction - oldDirection);

        return this.apiClient.post<ApiResponseBase>('comment/vote', {commenterToken: this.commenterTokenGet(), commentHex, direction})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    this.removeClasses(upvote, 'upvoted');
                    this.removeClasses(downvote, 'downvoted');
                    score.innerText = this.scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) - direction + oldDirection);
                    this.upDownOnClickSet(upvote, downvote, commentHex, oldDirection);
                    return;
                }
                this.errorHide();
            });
    }

    /**
     * Submit the entered comment markdown to the backend for saving.
     * @param commentHex Comment's hex ID
     */
    saveCommentEdits(commentHex: string): Promise<void> {
        const textarea = this.byId<HTMLTextAreaElement>(IDS.textarea + commentHex);
        const markdown = textarea.value.trim();
        if (markdown === '') {
            this.addClasses(textarea, 'red-border');
            return Promise.reject();
        }

        this.removeClasses(textarea, 'red-border');

        return this.apiClient.post<ApiCommentEditResponse>('comment/edit', {commenterToken: this.commenterTokenGet(), commentHex, markdown})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return;
                }

                this.errorHide();

                this.commentsByHex[commentHex].markdown = markdown;
                this.commentsByHex[commentHex].html = resp.html;

                // Hide the editor
                this.stopEditing(commentHex);

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
                    this.prepend(this.byId(IDS.superContainer + commentHex), this.messageCreate(message));
                }
            });
    }

    /**
     * Create a new editor for editing a comment with the given hex ID.
     * @param commentHex Comment's hex ID.
     */
    startEditing(commentHex: string) {
        if (this.shownEdit[commentHex]) {
            return;
        }

        const text = this.byId<HTMLDivElement>(IDS.text + commentHex);
        this.shownEdit[commentHex] = true;
        text.replaceWith(this.textareaCreate(commentHex, true));

        const textarea = this.byId<HTMLTextAreaElement>(IDS.textarea + commentHex);
        textarea.value = this.commentsByHex[commentHex].markdown;

        // Turn the Edit button into a Cancel edit button
        const editButton = this.byId<HTMLButtonElement>(IDS.edit + commentHex);
        this.removeClasses(editButton, 'option-edit');
        this.addClasses(editButton, 'option-cancel');
        editButton.title = 'Cancel edit';
        this.onClick(this.removeAllEventListeners(editButton), () => this.stopEditing(commentHex));
    }

    /**
     * Close the created editor for editing a comment with the given hex ID, cancelling the edits.
     * @param commentHex Comment's hex ID.
     */
    stopEditing(commentHex: string) {
        const cont = this.byId(IDS.superContainer + commentHex);
        cont.innerHTML = this.commentsByHex[commentHex].html;
        cont.id = Comentario.idPrefix + IDS.text + commentHex;
        delete this.shownEdit[commentHex];

        // Turn the Cancel edit button back into the Edit button
        const editButton = this.byId(IDS.edit + commentHex);
        this.addClasses(editButton, 'option-edit');
        this.removeClasses(editButton, 'option-cancel');
        editButton.title = 'Edit comment';

        // Bind comment editing to a click
        this.onClick(this.removeAllEventListeners(editButton), () => this.startEditing(commentHex));
    }

    /**
     * Create a new editor for editing a reply to the comment with the given hex ID.
     * @param commentHex Comment's hex ID.
     */
    replyShow(commentHex: string) {
        // Don't bother if there's an editor already
        if (this.shownReply[commentHex]) {
            return;
        }

        const text = this.byId(IDS.text + commentHex);
        this.insertAfter(text, this.textareaCreate(commentHex, false));
        this.shownReply[commentHex] = true;

        let replyButton = this.byId(IDS.reply + commentHex);

        this.removeClasses(replyButton, 'option-reply');
        this.addClasses(replyButton, 'option-cancel');

        replyButton.title = 'Cancel reply';

        replyButton = this.removeAllEventListeners(replyButton);
        this.onClick(replyButton, () => this.replyCollapse(commentHex));
    }

    /**
     * Close the created editor for editing a reply to the comment with the given hex ID.
     * @param commentHex Comment's hex ID.
     */
    replyCollapse(commentHex: string) {
        let replyButton = this.byId(IDS.reply + commentHex);
        const el = this.byId(IDS.superContainer + commentHex);

        el.remove();
        delete this.shownReply[commentHex];

        this.addClasses(replyButton, 'option-reply');
        this.removeClasses(replyButton, 'option-cancel');

        replyButton.title = 'Reply to this comment';

        replyButton = this.removeAllEventListeners(replyButton);
        this.onClick(replyButton, () => this.replyShow(commentHex));
    }

    commentCollapse(id: string) {
        const children = this.byId(IDS.children + id);
        if (children) {
            this.addClasses(children, 'hidden');
        }

        let button = this.byId(IDS.collapse + id);
        this.removeClasses(button, 'option-collapse');
        this.addClasses(button, 'option-uncollapse');

        button.title = 'Expand children';

        button = this.removeAllEventListeners(button);
        this.onClick(button, () => this.commentUncollapse(id));
    }

    commentUncollapse(id: string) {
        const children = this.byId(IDS.children + id);
        let button = this.byId(IDS.collapse + id);

        if (children) {
            this.removeClasses(children, 'hidden');
        }

        this.removeClasses(button, 'option-uncollapse');
        this.addClasses(button, 'option-collapse');

        button.title = 'Collapse children';

        button = this.removeAllEventListeners(button);
        this.onClick(button, () => this.commentCollapse(id));
    }

    parentMap(comments: Comment[]): CommentsMap {
        const m: CommentsMap = {};
        comments.forEach(comment => {
            const parentHex = comment.parentHex;
            if (!(parentHex in m)) {
                m[parentHex] = [];
            }

            comment.creationMs = new Date(comment.creationDate).getTime();

            m[parentHex].push(comment);
            this.commentsByHex[comment.commentHex] = {
                html: comment.html,
                markdown: comment.markdown,
            };
        });

        return m;
    }

    commentsRender() {
        const commentsArea = this.byId(IDS.commentsArea);
        commentsArea.innerHTML = '';

        const cards = this.commentsRecurse(this.parentMap(this.comments), 'root');
        if (cards) {
            this.append(commentsArea, cards);
        }
    }

    submitAuthenticated(id: string): Promise<void> {
        if (this.isAuthenticated) {
            return this.commentNew(id, this.commenterTokenGet(), true);
        }

        this.loginBoxShow(id);
        return Promise.resolve();
    }

    submitAnonymous(id: string): Promise<void> {
        this.chosenAnonymous = true;
        return this.commentNew(id, 'anonymous', true);
    }

    submitAccountDecide(id: string): Promise<void> {
        if (this.requireIdentification) {
            return this.submitAuthenticated(id);
        }

        const anonCheckbox = this.byId<HTMLInputElement>(IDS.anonymousCheckbox + id);
        const textarea = this.byId<HTMLTextAreaElement>(IDS.textarea + id);
        const markdown = textarea.value.trim();

        if (markdown === '') {
            this.addClasses(textarea, 'red-border');
            return Promise.reject();
        }

        this.removeClasses(textarea, 'red-border');
        return anonCheckbox.checked ? this.submitAnonymous(id) : this.submitAuthenticated(id);
    }

    // OAuth logic
    commentoAuth(provider: string, commentHex: string): Promise<void> {
        // Open a popup window
        const popup = window.open('', '_blank');

        // Request a token
        return this.apiClient.get<ApiCommenterTokenNewResponse>('commenter/token/new')
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return this.reject(resp.message);
                }

                this.errorHide();
                this.cookieSet('commentoCommenterToken', resp.commenterToken);
                popup.location = `${this.origin}/api/oauth/${provider}/redirect?commenterToken=${resp.commenterToken}`;

                // Wait until the popup is closed
                return new Promise<void>(resolve => {
                    const interval = setInterval(
                        () => {
                            if (popup.closed) {
                                clearInterval(interval);
                                resolve();
                            }
                        },
                        250);
                });
            })
            // Refresh the auth status
            .then(() => this.selfGet())
            // Update the login controls
            .then(() => {
                this.setAttr(this.byId(IDS.loggedContainer), {style: null});

                // Hide the login button
                this.remove(this.byId(IDS.login));

                // Submit the pending comment, if there was one
                return commentHex && this.commentNew(commentHex, this.commenterTokenGet(), false);
            })
            .then(() => this.loginBoxClose())
            .then(() => this.commentsGet())
            .then(() => this.commentsRender());
    }

    loginBoxCreate() {
        this.append(this.root, this.createElement('div', {id: IDS.loginBoxContainer}));
    }

    popupRender(commentHex: string) {
        const loginBoxContainer = this.byId(IDS.loginBoxContainer);
        this.addClasses(loginBoxContainer, 'login-box-container');
        this.setAttr(loginBoxContainer, {style: 'display: none; opacity: 0;'});

        const loginBox = this.createElement('form', {id: IDS.loginBox, classes: 'login-box'});
        // This is ugly, must redesign the whole bloody login/signup form
        loginBox.addEventListener('submit', (e) => {
            e.preventDefault();
            if (!this.byId<HTMLButtonElement>(IDS.loginBoxPasswordButton)) {
                this.showPasswordField();
            } else if (this.popupBoxType === 'login') {
                this.login(commentHex);
            } else {
                this.signup(commentHex);
            }
        });

        const ssoSubtitle           = this.createElement('div',    {id: IDS.loginBoxSsoPretext, classes: 'login-box-subtitle', innerText: `Proceed with ${parent.location.host} authentication`});
        const ssoButtonContainer    = this.createElement('div',    {id: IDS.loginBoxSsoButtonContainer, classes: 'oauth-buttons-container'});
        const ssoButton             = this.createElement('div',    {classes: 'oauth-buttons'});
        const hr1                   = this.createElement('hr',     {id: IDS.loginBoxHr1});
        const oauthSubtitle         = this.createElement('div',    {id: IDS.loginBoxOauthPretext, classes: 'login-box-subtitle', innerText: 'Proceed with social login'});
        const oauthButtonsContainer = this.createElement('div',    {id: IDS.loginBoxOauthButtonsContainer, classes: 'oauth-buttons-container'});
        const oauthButtons          = this.createElement('div',    {classes: 'oauth-buttons'});
        const hr2                   = this.createElement('hr',     {id: IDS.loginBoxHr2});
        const emailSubtitle         = this.createElement('div',    {id: IDS.loginBoxEmailSubtitle, classes: 'login-box-subtitle', innerText: 'Login with your email address'});
        const emailButton           = this.createElement('button', {id: IDS.loginBoxEmailButton, type: 'submit', classes: 'email-button', innerText: 'Continue'});
        const emailContainer        = this.createElement('div', {
            classes: 'email-container',
            children: [
                this.createElement('div', {
                    classes:  'email',
                    children: [
                        this.createElement('input', {
                            id:           IDS.loginBoxEmailInput,
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
        const forgotLinkContainer = this.createElement('div', {id: IDS.loginBoxForgotLinkContainer, classes: 'forgot-link-container'});
        const forgotLink          = this.createElement('a',   {classes: 'forgot-link', innerText: 'Forgot your password?', parent: forgotLinkContainer});
        const loginLinkContainer  = this.createElement('div', {id: IDS.loginBoxLoginLinkContainer, classes: 'login-link-container'});
        const loginLink           = this.createElement('a',   {classes: 'login-link', innerText: 'Don\'t have an account? Sign up.', parent: loginLinkContainer});
        const close               = this.createElement('div', {classes: 'login-box-close', parent: loginBox});

        this.addClasses(this.root, 'root-min-height');

        this.onClick(forgotLink,  () => this.forgotPassword());
        this.onClick(loginLink,   () => this.popupSwitch());
        this.onClick(close,       () => this.loginBoxClose());

        let hasOAuth = false;
        const oauthProviders = ['google', 'github', 'gitlab'];
        oauthProviders.filter(p => this.configuredOauths[p]).forEach(provider => {
            const button = this.createElement(
                'button',
                {classes: ['button', `${provider}-button`], type: 'button', innerText: provider, parent: oauthButtons});
            this.onClick(button, () => this.commentoAuth(provider, commentHex));
            hasOAuth = true;
        });

        if (this.configuredOauths['sso']) {
            const button = this.createElement(
                'button',
                {classes: ['button', 'sso-button'], type: 'button', innerText: 'Single Sign-On', parent: ssoButton});
            this.onClick(button, () => this.commentoAuth('sso', commentHex));
            this.append(ssoButtonContainer, ssoButton);
            this.append(loginBox, ssoSubtitle);
            this.append(loginBox, ssoButtonContainer);

            if (hasOAuth || this.configuredOauths['commento']) {
                this.append(loginBox, hr1);
            }
        }

        this.oauthButtonsShown = hasOAuth;
        if (hasOAuth) {
            this.append(oauthButtonsContainer, oauthButtons);
            this.append(loginBox, oauthSubtitle, oauthButtonsContainer);
            if (this.configuredOauths['commento']) {
                this.append(loginBox, hr2);
            }
        }

        if (this.configuredOauths['commento']) {
            this.append(loginBox, emailSubtitle, emailContainer, forgotLinkContainer, loginLinkContainer);
        }

        this.popupBoxType = 'login';
        loginBoxContainer.innerHTML = '';
        this.append(loginBoxContainer, loginBox);
    }

    forgotPassword() {
        const popup = window.open('', '_blank');
        popup.location = `${this.origin}/forgot?commenter=true`;
        this.loginBoxClose();
    }

    popupSwitch() {
        const emailSubtitle = this.byId(IDS.loginBoxEmailSubtitle);

        if (this.oauthButtonsShown) {
            this.remove(
                this.byId(IDS.loginBoxOauthButtonsContainer),
                this.byId(IDS.loginBoxOauthPretext),
                this.byId(IDS.loginBoxHr1),
                this.byId(IDS.loginBoxHr2));
        }

        if (this.configuredOauths['sso']) {
            this.remove(
                this.byId(IDS.loginBoxSsoButtonContainer),
                this.byId(IDS.loginBoxSsoPretext),
                this.byId(IDS.loginBoxHr1),
                this.byId(IDS.loginBoxHr2));
        }

        this.remove(this.byId(IDS.loginBoxLoginLinkContainer), this.byId(IDS.loginBoxForgotLinkContainer));

        emailSubtitle.innerText = 'Create an account';
        this.popupBoxType = 'signup';
        this.showPasswordField();
        this.byId(IDS.loginBoxEmailInput).focus();
    }

    loginUP(email: string, password: string, id: string): Promise<void> {
        return this.apiClient.post<ApiCommenterLoginResponse>('commenter/login', {email, password})
            .then(resp => {
                if (!resp.success) {
                    this.loginBoxClose();
                    this.errorShow(resp.message);
                    return Promise.reject();
                }

                this.errorHide();
                this.cookieSet('commentoCommenterToken', resp.commenterToken);
                this.selfLoad(resp.commenter, resp.email);
                this.remove(this.byId(IDS.login));
                return (id ? this.commentNew(id, resp.commenterToken, false) : undefined);
            })
            .then(() => this.loginBoxClose())
            .then(() => this.commentsGet())
            .then(() => this.commentsRender())
            .then(() => this.allShow());
    }

    login(id: string): Promise<void> {
        const email    = this.byId<HTMLInputElement>(IDS.loginBoxEmailInput);
        const password = this.byId<HTMLInputElement>(IDS.loginBoxPasswordInput);
        return this.loginUP(email.value, password.value, id);
    }

    signup(id: string): Promise<void> {
        const email    = this.byId<HTMLInputElement>(IDS.loginBoxEmailInput);
        const name     = this.byId<HTMLInputElement>(IDS.loginBoxNameInput);
        const website  = this.byId<HTMLInputElement>(IDS.loginBoxWebsiteInput);
        const password = this.byId<HTMLInputElement>(IDS.loginBoxPasswordInput);

        const data = {
            email:    email.value,
            name:     name.value,
            website:  website.value,
            password: password.value,
        };

        return this.apiClient.post<ApiResponseBase>('commenter/new', data)
            .then(resp => {
                if (!resp.success) {
                    this.loginBoxClose();
                    this.errorShow(resp.message);
                    return Promise.reject();
                }

                this.errorHide();
                return undefined;
            })
            .then(() => this.loginUP(data.email, data.password, id));
    }

    showPasswordField() {
        const isSignup = this.popupBoxType === 'signup';
        const loginBox = this.byId(IDS.loginBox);
        const subtitle = this.byId(IDS.loginBoxEmailSubtitle);

        this.remove(
            this.byId(IDS.loginBoxEmailButton),
            this.byId(IDS.loginBoxLoginLinkContainer),
            this.byId(IDS.loginBoxForgotLinkContainer));
        if (this.oauthButtonsShown && Object.keys(this.configuredOauths).length) {
            this.remove(
                this.byId(IDS.loginBoxHr1),
                this.byId(IDS.loginBoxHr2),
                this.byId(IDS.loginBoxOauthPretext),
                this.byId(IDS.loginBoxOauthButtonsContainer));
        }

        const controls = isSignup ?
            [
                {id: IDS.loginBoxNameInput,     classes: 'input', name: 'name',     type: 'text',     placeholder: 'Real Name'},
                {id: IDS.loginBoxWebsiteInput,  classes: 'input', name: 'website',  type: 'text',     placeholder: 'Website (Optional)'},
                {id: IDS.loginBoxPasswordInput, classes: 'input', name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'new-password'},
            ] :
            [
                {id: IDS.loginBoxPasswordInput, classes: 'input', name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'current-password'},
            ];

        subtitle.innerText = isSignup ?
            'Finish the rest of your profile to complete.' :
            'Enter your password to log in.';

        controls.forEach(c => {
            const fieldContainer = this.createElement('div', {classes: 'email-container'});
            const field          = this.createElement('div', {classes: 'email', parent: fieldContainer});
            const fieldInput     = this.createElement('input', c);
            this.append(field, fieldInput);
            // Add a submit button next to the password input
            if (c.type === 'password') {
                this.createElement('button', {
                    id:        IDS.loginBoxPasswordButton,
                    type:      'submit',
                    classes:   'email-button',
                    innerText: this.popupBoxType,
                    parent:    field,
                });
            }
            this.append(loginBox, fieldContainer);
        });

        this.byId(isSignup ? IDS.loginBoxNameInput : IDS.loginBoxPasswordInput).focus();
    }

    pageUpdate(): Promise<void> {
        const data = {
            commenterToken: this.commenterTokenGet(),
            domain:         parent.location.host,
            path:           this.pageId,
            attributes:     {isLocked: this.isLocked, stickyCommentHex: this.stickyCommentHex},
        };

        return this.apiClient.post<ApiResponseBase>('page/update', data)
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return Promise.reject();
                }

                this.errorHide();
                return undefined;
            });
    }

    threadLockToggle(): Promise<void> {
        const lock = this.byId<HTMLButtonElement>(IDS.modToolsLockButton);
        this.isLocked = !this.isLocked;
        lock.disabled = true;
        return this.pageUpdate()
            .then(() => lock.disabled = false)
            .then(() => this.reload());
    }

    commentSticky(commentHex: string): Promise<void> {
        if (this.stickyCommentHex !== 'none') {
            const sticky = this.byId(IDS.sticky + this.stickyCommentHex);
            this.removeClasses(sticky, 'option-unsticky');
            this.addClasses(sticky, 'option-sticky');
        }

        this.stickyCommentHex = this.stickyCommentHex === commentHex ? 'none' : commentHex;

        return this.pageUpdate()
            .then(() => {
                const sticky = this.byId(IDS.sticky + commentHex);
                if (this.stickyCommentHex === commentHex) {
                    this.removeClasses(sticky, 'option-sticky');
                    this.addClasses(sticky, 'option-unsticky');
                } else {
                    this.removeClasses(sticky, 'option-unsticky');
                    this.addClasses(sticky, 'option-sticky');
                }
            });
    }

    mainAreaCreate() {
        this.createElement('div', {id: IDS.mainArea, classes: 'main-area', style: 'display: none', parent: this.root});
    }

    modToolsCreate() {
        const btnLock = this.createElement(
            'button',
            {
                id:        IDS.modToolsLockButton,
                type:      'button',
                innerHTML: this.isLocked ? 'Unlock Thread' : 'Lock Thread',
            });
        this.onClick(btnLock, this.threadLockToggle);
        this.createElement(
            'div',
            {id: IDS.modTools, classes: 'mod-tools', style: 'display: none', parent: this.root, children: [btnLock]});
    }

    allShow() {
        const mainArea = this.byId(IDS.mainArea);
        const modTools = this.byId(IDS.modTools);
        const loggedContainer = this.byId(IDS.loggedContainer);

        this.setAttr(mainArea, {style: null});

        if (this.isModerator) {
            this.setAttr(modTools, {style: null});
        }

        if (loggedContainer) {
            this.setAttr(loggedContainer, {style: null});
        }
    }

    loginBoxClose() {
        const mainArea = this.byId(IDS.mainArea);
        const loginBoxContainer = this.byId(IDS.loginBoxContainer);

        this.removeClasses(mainArea, 'blurred');
        this.removeClasses(this.root, 'root-min-height');

        this.setAttr(loginBoxContainer, {style: 'display: none'});
    }

    loginBoxShow(commentHex: string) {
        const mainArea = this.byId(IDS.mainArea);
        const loginBoxContainer = this.byId(IDS.loginBoxContainer);

        this.popupRender(commentHex);

        this.addClasses(mainArea, 'blurred');
        this.setAttr(loginBoxContainer, {style: null});

        loginBoxContainer.scrollIntoView({behavior: 'smooth'});

        this.byId(IDS.loginBoxEmailInput).focus();
    }

    dataTagsLoad() {
        for (const script of this.doc.getElementsByTagName('script')) {
            if (script.src.match(/\/js\/commento\.js$/)) {
                let s = this.getAttr(script, 'data-page-id');
                if (s) {
                    this.pageId = s;
                }
                this.cssOverride = this.getAttr(script, 'data-css-override');
                this.autoInit = this.getAttr(script, 'data-auto-init') !== 'false';
                s = this.getAttr(script, 'data-id-root');
                if (s) {
                    this.rootId = s;
                }
                this.noFonts = this.getAttr(script, 'data-no-fonts') === 'true';
                this.hideDeleted = this.getAttr(script, 'data-hide-deleted') === 'true';
                break;
            }
        }
    }

    loadHash() {
        if (window.location.hash) {
            if (window.location.hash.startsWith('#commento-')) {
                const id = window.location.hash.split('-')[1];
                const el = this.byId(IDS.card + id);
                if (el === null) {
                    if (id.length === 64) {
                        // A hack to make sure it's a valid ID before showing the user a message.
                        this.errorShow('The comment you\'re looking for no longer exists or was deleted.');
                    }
                    return;
                }

                this.addClasses(el, 'highlighted-card');
                el.scrollIntoView(true);
            } else if (window.location.hash.startsWith('#commento')) {
                this.root.scrollIntoView(true);
            }
        }
    }
}
