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
import { Wrap } from './element-wrap';

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

    private readonly origin = '[[[.Origin]]]';
    private readonly cdn = '[[[.CdnPrefix]]]';

    /** HTTP client we'll use for API requests. */
    private readonly apiClient = new HttpClient(`${this.origin}/api`);

    /** Default ID of the container element Comentario will be embedded into. */
    private rootId = 'commento';

    private root: Wrap<any>;
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
        this.root = Wrap.byId(this.rootId, true);
        if (!this.root.ok) {
            return this.reject(`No root element with id='${this.rootId}' found. Check your configuration and HTML.`);
        }

        // TODO refactor this
        //if (this.mobileView === null) {
        //    this.mobileView = this.root.getBoundingClientRect()['width'] < 450;
        //}

        this.root.classes('root', !this.noFonts && 'root-font');

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
        Wrap.byId(this.rootId, true).html('');
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
                this.root.append(this.footerLoad());
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

        const loggedContainer = Wrap.new('div').id(IDS.loggedContainer).classes('logged-container').style('display: none');
        const loggedInAs      = Wrap.new('div').classes('logged-in-as').appendTo(loggedContainer);
        const name            = Wrap.new(commenter.link !== 'undefined' ? 'a' : 'div').classes('name').inner(commenter.name).appendTo(loggedInAs);
        const btnSettings     = Wrap.new('div').classes('profile-button').inner('Notification Settings').click(() => this.notificationSettings(email.unsubscribeSecretHex));
        const btnEditProfile  = Wrap.new('div').classes('profile-button').inner('Edit Profile').click(() => this.profileEdit());
        Wrap.new('div').classes('profile-button').inner('Logout').click(() => this.logout()).appendTo(loggedContainer);
        const color = this.colorGet(`${commenter.commenterHex}-${commenter.name}`);

        // Set the profile href for the commenter, if any
        if (commenter.link !== 'undefined') {
            name.attr({href: commenter.link});
        }

        // Add an avatar
        if (commenter.photo === 'undefined') {
            Wrap.new('div')
                .classes('avatar')
                .html(commenter.name[0].toUpperCase())
                .style(`background-color: ${color}`)
                .appendTo(loggedInAs);
        } else {
            Wrap.new('img')
                .classes('avatar-img')
                .attr({src: `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`, loading: 'lazy', alt: ''})
                .appendTo(loggedInAs);
        }

        // If it's a local user, add an Edit profile button
        if (commenter.provider === 'commento') {
            loggedContainer.append(btnEditProfile);
        }
        loggedContainer.append(btnSettings);

        // Add the container to the root
        loggedContainer.prependTo(this.root);
        this.isAuthenticated = true;
    }

    selfGet(): Promise<void> {
        const commenterToken = this.commenterTokenGet();
        if (commenterToken === 'anonymous') {
            this.isAuthenticated = false;
            return Promise.resolve();
        }

        return this.apiClient.post<ApiSelfResponse>('commenter/self', {commenterToken: this.commenterTokenGet()})
            // On any error consider the user unauthenticated
            .catch(() => null)
            .then(resp => {
                if (!resp?.success) {
                    this.cookieSet('commentoCommenterToken', 'anonymous');
                    return;
                }

                this.selfLoad(resp.commenter, resp.email);
                this.allShow();
                return;
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
                new Wrap(this.doc.getElementsByTagName('head')[0])
                    .append(
                        Wrap.new('link').attr({href: url, rel: 'stylesheet', type: 'text/css'}).load(resolve));
            });
    }

    footerLoad(): Wrap<HTMLDivElement> {
        return Wrap.new('div')
            .id(IDS.footer)
            .classes('footer')
            .append(
                Wrap.new('div')
                    .classes('logo-container')
                    .append(
                        Wrap.new('a')
                            .classes('logo')
                            .attr({href: 'https://comentario.app/', target: '_blank'})
                            .append(Wrap.new('span').classes('logo-text').inner('Comentario 🗨'))));
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
        Wrap.byId(IDS.error).inner(text).style('display: block;');
    }

    errorHide() {
        Wrap.byId(IDS.error).style('display: none;');
    }

    errorElementCreate() {
        Wrap.new('div').id(IDS.error).classes('error-box').style('display: none;').append(this.root);
    }

    autoExpander(el: HTMLElement): CallbackFunction {
        return () => {
            el.style.height = '';
            el.style.height = `${Math.min(Math.max(el.scrollHeight, 75), 400)}px`;
        };
    }

    markdownHelpShow(commentHex: string) {
        Wrap.new('table')
            .id(IDS.markdownHelp + commentHex)
            .classes('markdown-help')
            .appendTo(Wrap.byId(IDS.superContainer + commentHex))
            .append(
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<i>italics</i>'),
                        Wrap.new('td').html('surround text with <pre>*asterisks*</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<b>bold</b>'),
                        Wrap.new('td').html('surround text with <pre>**two asterisks**</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<pre>code</pre>'),
                        Wrap.new('td').html('surround text with <pre>`backticks`</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<del>strikethrough</del>'),
                        Wrap.new('td').html('surround text with <pre>~~two tilde characters~~</pre>')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<a href="https://example.com">hyperlink</a>'),
                        Wrap.new('td').html('<pre>[hyperlink](https://example.com)</pre> or just a bare URL')),
                Wrap.new('tr')
                    .append(
                        Wrap.new('td').html('<blockquote>quote</blockquote>'),
                        Wrap.new('td').html('prefix with <pre>&gt;</pre>')));

        // Add a collapse button
        Wrap.byId(IDS.markdownButton + commentHex).unlisten().click(() => this.markdownHelpHide(commentHex));
    }

    markdownHelpHide(commentHex: string) {
        Wrap.byId(IDS.markdownButton + commentHex).unlisten().click(() => this.markdownHelpShow(commentHex));
        Wrap.byId(IDS.markdownHelp + commentHex).remove();
    }

    /**
     * Create a new editor for editing comment text.
     * @param commentHex Comment's hex ID.
     * @param isEdit Whether it's adding a new comment (false) or editing an existing one (true)
     */
    textareaCreate(commentHex: string, isEdit: boolean): Wrap<HTMLDivElement> {
        const textOuter = Wrap.new('div').id(IDS.superContainer + commentHex).classes('button-margin')
            .append(
                // Text area in a container
                Wrap.new('div').id(IDS.textareaContainer + commentHex).classes('textarea-container')
                    .append(
                        Wrap.new('textarea').id(IDS.textarea + commentHex).attr({placeholder: 'Add a comment'}).autoExpand()),
                // Save button
                Wrap.new('button')
                    .id(IDS.submitButton + commentHex)
                    .attr({type: 'submit'})
                    .classes('button', 'submit-button')
                    .inner(isEdit ? 'Save Changes' : 'Add Comment')
                    .click(() => isEdit ? this.saveCommentEdits(commentHex) : this.submitAccountDecide(commentHex)));

        // "Comment anonymously" checkbox
        const anonCheckbox = Wrap.new('input').id(IDS.anonymousCheckbox + commentHex).attr({type: 'checkbox'});
        if (this.anonymousOnly) {
            anonCheckbox.checked(true).attr({disabled: 'true'});
        }
        const anonCheckboxCont = Wrap.new('div')
            .classes('round-check', 'anonymous-checkbox-container')
            .append(
                anonCheckbox,
                Wrap.new('label').attr({for: Wrap.idPrefix + IDS.anonymousCheckbox + commentHex}).inner('Comment anonymously'));

        if (!this.requireIdentification && !isEdit) {
            textOuter.append(anonCheckboxCont);
        }

        // Markdown help button
        Wrap.new('a')
            .id(IDS.markdownButton + commentHex)
            .classes('markdown-button')
            .html('<b>M⬇</b>&nbsp;Markdown')
            .click(() => this.markdownHelpShow(commentHex))
            .appendTo(textOuter);
        return textOuter;
    }

    sortPolicyApply(policy: SortPolicy) {
        Wrap.byId(IDS.sortPolicy + this.sortPolicy).noClasses('sort-policy-button-selected');

        const commentsArea = Wrap.byId(IDS.commentsArea);
        commentsArea.innerHTML = '';
        this.sortPolicy = policy;
        const cards = this.commentsRecurse(this.parentMap(this.comments), 'root');
        if (cards) {
            commentsArea.append(cards);
        }

        this.addClasses(Wrap.byId(IDS.sortPolicy + policy), 'sort-policy-button-selected');
    }

    sortPolicyBox(): Wrap<HTMLDivElement> {
        const container = Wrap.new('div').classes('sort-policy-buttons-container');
        const buttonBar = Wrap.new('div').classes('sort-policy-buttons').appendTo(container);
        Object.keys(this.sortingProps).forEach((sp: SortPolicy) =>
            Wrap.new('a')
                .id(IDS.sortPolicy + sp)
                .classes('sort-policy-button', sp === this.sortPolicy && 'sort-policy-button-selected')
                .inner(this.sortingProps[sp].label)
                .appendTo(buttonBar)
                .click(() => this.sortPolicyApply(sp)));
        return container;
    }

    /**
     * Create the top-level ("main area") elements in the root.
     */
    rootCreate(): void {
        const mainArea = Wrap.byId(IDS.mainArea);
        const login           = Wrap.new('div').id(IDS.login).classes('login');
        const loginText       = Wrap.new('div').classes('login-text').inner('Login').click(() => this.loginBoxShow(null));
        const preCommentsArea = Wrap.new('div').id(IDS.preCommentsArea);
        const commentsArea    = Wrap.new('div').id(IDS.commentsArea).classes('comments');

        // If there's an OAuth provider configured, add a Login button
        if (Object.keys(this.configuredOauths).some(k => this.configuredOauths[k])) {
            login.append(loginText);
        } else if (!this.requireIdentification) {
            this.anonymousOnly = true;
        }

        if (this.isLocked || this.isFrozen) {
            if (this.isAuthenticated || this.chosenAnonymous) {
                mainArea.append(this.messageCreate('This thread is locked. You cannot add new comments.'));
                login.remove();
            } else {
                // Add a root editor (for creating a new comment)
                mainArea.append(login, this.textareaCreate('root', false));
            }
        } else {
            if (this.isAuthenticated) {
                login.remove();
            } else {
                mainArea.append(login);
            }
            // Add a root editor (for creating a new comment)
            mainArea.append(this.textareaCreate('root', false));
        }

        // If there's any comment, add sort buttons
        if (this.comments.length) {
            mainArea.append(this.sortPolicyBox());
        }
        mainArea.append(preCommentsArea, commentsArea).appendTo(this.root);
    }

    messageCreate(text: string): Wrap<HTMLDivElement> {
        return Wrap.new('div').classes('moderation-notice').inner(text);
    }

    commentNew(commentHex: string, commenterToken: string, appendCard: boolean): Promise<void> {
        const container   = Wrap.byId(IDS.superContainer + commentHex);
        const textarea    = Wrap.byId(IDS.textarea + commentHex);
        const replyButton = Wrap.byId(IDS.reply + commentHex);

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
                    this.prepend(Wrap.byId(IDS.superContainer + commentHex), this.messageCreate(message));
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
                        this.insertAfter(Wrap.byId(IDS.preCommentsArea), newCard);
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

    commentsRecurse(parentMap: CommentsMap, parentHex: string): Wrap<HTMLDivElement> {
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
        const cards = Wrap.new('div');
        cur.forEach(comment => {
            const commenter = this.commenters[comment.commenterHex];
            const hex = comment.commentHex;
            const header = Wrap.new('div').classes('header');
            const name = Wrap.new(commenter.link !== 'undefined' && commenter.link !== 'https://undefined' && commenter.link !== '' ? 'a' : 'div')
                .id(IDS.name + hex)
                .inner(comment.deleted ? '[deleted]' : commenter.name)
                .classes('name');
            const color = this.colorGet(`${comment.commenterHex}-${commenter.name}`);
            const card     = Wrap.new('div').id(IDS.card     + hex).style(`border-left: 2px solid ${color}`).classes('card');
            const subtitle = Wrap.new('div').id(IDS.subtitle + hex).classes('subtitle');
            const timeago = Wrap.new('div')
                .id(IDS.timeago + hex)
                .classes('timeago')
                .html(this.timeDifference(curTime, comment.creationMs))
                .attr({title: comment.creationDate.toString()});
            const score    = Wrap.new('div')   .id(IDS.score    + hex).classes('score').inner(this.scorify(comment.score));
            const body     = Wrap.new('div')   .id(IDS.body     + hex).classes('body');
            const text     = Wrap.new('div')   .id(IDS.text     + hex).html(comment.html);
            const options  = Wrap.new('div')   .id(IDS.options  + hex).classes('options');
            const edit     = Wrap.new('button').id(IDS.edit     + hex).classes('option-button', 'option-edit')    .attr({type: 'button', title: 'Edit'}).click(() => this.startEditing(hex));
            const reply    = Wrap.new('button').id(IDS.reply    + hex).classes('option-button', 'option-reply')   .attr({type: 'button', title: 'Reply'});
            const collapse = Wrap.new('button').id(IDS.collapse + hex).classes('option-button', 'option-collapse').attr({type: 'button', title: 'Collapse children'}).click(() => this.commentCollapse(hex));
            let   upvote   = Wrap.new('button').id(IDS.upvote   + hex).classes('option-button', 'option-upvote')  .attr({type: 'button', title: 'Upvote'});
            let   downvote = Wrap.new('button').id(IDS.downvote + hex).classes('option-button', 'option-downvote').attr({type: 'button', title: 'Downvote'});
            const approve  = Wrap.new('button').id(IDS.approve  + hex).classes('option-button', 'option-approve') .attr({type: 'button', title: 'Approve'}).click(() => this.commentApprove(hex));
            const remove   = Wrap.new('button').id(IDS.remove   + hex).classes('option-button', 'option-remove')  .attr({type: 'button', title: 'Remove'}).click(() => this.commentDelete(hex));
            const sticky   = Wrap.new('button')
                .id(IDS.sticky + hex)
                .classes('option-button', this.stickyCommentHex === hex ? 'option-unsticky' : 'option-sticky')
                .attr({
                    title: this.stickyCommentHex === hex ? this.isModerator ? 'Unsticky' : 'This comment has been stickied' : 'Sticky',
                    type: 'button',
                })
                .click(() => this.commentSticky(hex));
            const contents = Wrap.new('div').id(IDS.contents + hex);

            // TODO refactor
            // if (this.mobileView) {
            //     this.addClasses(options, 'options-mobile');
            // }

            const children = this.commentsRecurse(parentMap, hex).id(IDS.children + hex);

            let avatar;
            if (commenter.photo === 'undefined') {
                avatar = Wrap.new('div')
                    .style(`background-color: ${color}`)
                    .classes('avatar')
                    .html(comment.commenterHex === 'anonymous' ? '?' : commenter.name[0].toUpperCase());
            } else {
                avatar = Wrap.new('img')
                    .classes('avatar-img')
                    .attr({src: `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`, alt: ''});
            }
            if (this.isModerator && comment.state !== 'approved') {
                card.classes('dark-card');
            }
            name.classes(
                commenter.isModerator && 'moderator',
                comment.state === 'flagged' && 'flagged');

            if (this.isAuthenticated) {
                if (comment.direction > 0) {
                    upvote.classes('upvoted');
                } else if (comment.direction < 0) {
                    downvote.classes('downvoted');
                }
            }

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
                Wrap.new('div'), {classes: 'options-clearfix', parent: contents});
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

                const card = Wrap.byId(IDS.card + commentHex);
                const name = Wrap.byId(IDS.name + commentHex);
                const tick = Wrap.byId(IDS.approve + commentHex);

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
                const text = Wrap.byId(IDS.text + commentHex);
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
        const [upvote, downvote] = this.upDownOnClickSet(
            Wrap.byId(IDS.upvote   + commentHex),
            Wrap.byId(IDS.downvote + commentHex),
            commentHex,
            direction);

        this.removeClasses(upvote, 'upvoted');
        this.removeClasses(downvote, 'downvoted');
        if (direction > 0) {
            this.addClasses(upvote, 'upvoted');
        } else if (direction < 0) {
            this.addClasses(downvote, 'downvoted');
        }

        const score  = Wrap.byId(IDS.score + commentHex);
        score.innerText = this.scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) + direction - oldDirection);

        return this.apiClient.post<ApiResponseBase>('comment/vote', {commenterToken: this.commenterTokenGet(), commentHex, direction})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    this.removeClasses(upvote, 'upvoted');
                    this.removeClasses(downvote, 'downvoted');
                    score.innerText = this.scorify(parseInt(score.innerText.replace(/[^\d-.]/g, '')) - direction + oldDirection);
                    this.upDownOnClickSet(upvote, downvote, commentHex, oldDirection);
                    return Promise.reject();
                }
                this.errorHide();
                return undefined;
            });
    }

    /**
     * Submit the entered comment markdown to the backend for saving.
     * @param commentHex Comment's hex ID
     */
    saveCommentEdits(commentHex: string): Promise<void> {
        const textarea = Wrap.byId(IDS.textarea + commentHex);
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
                    this.prepend(Wrap.byId(IDS.superContainer + commentHex), this.messageCreate(message));
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

        const text = Wrap.byId(IDS.text + commentHex);
        this.shownEdit[commentHex] = true;
        text.replaceWith(this.textareaCreate(commentHex, true));

        const textarea = Wrap.byId(IDS.textarea + commentHex);
        textarea.value = this.commentsByHex[commentHex].markdown;

        // Turn the Edit button into a Cancel edit button
        const editButton = Wrap.byId(IDS.edit + commentHex);
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
        const cont = Wrap.byId(IDS.superContainer + commentHex);
        cont.innerHTML = this.commentsByHex[commentHex].html;
        cont.id = Comentario.idPrefix + IDS.text + commentHex;
        delete this.shownEdit[commentHex];

        // Turn the Cancel edit button back into the Edit button
        const editButton = Wrap.byId(IDS.edit + commentHex);
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

        const text = Wrap.byId(IDS.text + commentHex);
        this.insertAfter(text, this.textareaCreate(commentHex, false));
        this.shownReply[commentHex] = true;

        let replyButton = Wrap.byId(IDS.reply + commentHex);

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
        let replyButton = Wrap.byId(IDS.reply + commentHex);
        const el = Wrap.byId(IDS.superContainer + commentHex);

        el.remove();
        delete this.shownReply[commentHex];

        this.addClasses(replyButton, 'option-reply');
        this.removeClasses(replyButton, 'option-cancel');

        replyButton.title = 'Reply to this comment';

        replyButton = this.removeAllEventListeners(replyButton);
        this.onClick(replyButton, () => this.replyShow(commentHex));
    }

    commentCollapse(id: string) {
        const children = Wrap.byId(IDS.children + id);
        if (children) {
            this.addClasses(children, 'hidden');
        }

        let button = Wrap.byId(IDS.collapse + id);
        this.removeClasses(button, 'option-collapse');
        this.addClasses(button, 'option-uncollapse');

        button.title = 'Expand children';

        button = this.removeAllEventListeners(button);
        this.onClick(button, () => this.commentUncollapse(id));
    }

    commentUncollapse(id: string) {
        const children = Wrap.byId(IDS.children + id);
        let button = Wrap.byId(IDS.collapse + id);

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
        const commentsArea = Wrap.byId(IDS.commentsArea);
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

        const anonCheckbox = Wrap.byId(IDS.anonymousCheckbox + id);
        const textarea = Wrap.byId(IDS.textarea + id);
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
                this.setAttr(Wrap.byId(IDS.loggedContainer), {style: null});

                // Hide the login button
                this.remove(Wrap.byId(IDS.login));

                // Submit the pending comment, if there was one
                return commentHex && this.commentNew(commentHex, this.commenterTokenGet(), false);
            })
            .then(() => this.loginBoxClose())
            .then(() => this.commentsGet())
            .then(() => this.commentsRender());
    }

    loginBoxCreate() {
        Wrap.new('div').id(IDS.loginBoxContainer).appendTo(this.root);
    }

    popupRender(commentHex: string) {
        const loginBoxContainer = Wrap.byId(IDS.loginBoxContainer);
        this.addClasses(loginBoxContainer, 'login-box-container');
        this.setAttr(loginBoxContainer, {style: 'display: none; opacity: 0;'});

        const loginBox = Wrap.new('form'), {id: IDS.loginBox, classes: 'login-box'});
        // This is ugly, must redesign the whole bloody login/signup form
        loginBox.addEventListener('submit', (e) => {
            e.preventDefault();
            if (!Wrap.byId(IDS.loginBoxPasswordButton)) {
                this.showPasswordField();
            } else if (this.popupBoxType === 'login') {
                this.login(commentHex);
            } else {
                this.signup(commentHex);
            }
        });

        const ssoSubtitle           = Wrap.new('div'),    {id: IDS.loginBoxSsoPretext, classes: 'login-box-subtitle', innerText: `Proceed with ${parent.location.host} authentication`});
        const ssoButtonContainer    = Wrap.new('div'),    {id: IDS.loginBoxSsoButtonContainer, classes: 'oauth-buttons-container'});
        const ssoButton             = Wrap.new('div'),    {classes: 'oauth-buttons'});
        const hr1                   = Wrap.new('hr'),     {id: IDS.loginBoxHr1});
        const oauthSubtitle         = Wrap.new('div'),    {id: IDS.loginBoxOauthPretext, classes: 'login-box-subtitle', innerText: 'Proceed with social login'});
        const oauthButtonsContainer = Wrap.new('div'),    {id: IDS.loginBoxOauthButtonsContainer, classes: 'oauth-buttons-container'});
        const oauthButtons          = Wrap.new('div'),    {classes: 'oauth-buttons'});
        const hr2                   = Wrap.new('hr'),     {id: IDS.loginBoxHr2});
        const emailSubtitle         = Wrap.new('div'),    {id: IDS.loginBoxEmailSubtitle, classes: 'login-box-subtitle', innerText: 'Login with your email address'});
        const emailButton           = Wrap.new('button'), {id: IDS.loginBoxEmailButton, type: 'submit', classes: 'email-button', innerText: 'Continue'});
        const emailContainer        = Wrap.new('div'), {
            classes: 'email-container',
            children: [
                Wrap.new('div'), {
                    classes:  'email',
                    .append(
                        Wrap.new('input'), {
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
        const forgotLinkContainer = Wrap.new('div'), {id: IDS.loginBoxForgotLinkContainer, classes: 'forgot-link-container'});
        const forgotLink          = Wrap.new('a'),   {classes: 'forgot-link', innerText: 'Forgot your password?', parent: forgotLinkContainer});
        const loginLinkContainer  = Wrap.new('div'), {id: IDS.loginBoxLoginLinkContainer, classes: 'login-link-container'});
        const loginLink           = Wrap.new('a'),   {classes: 'login-link', innerText: 'Don\'t have an account? Sign up.', parent: loginLinkContainer});
        const close               = Wrap.new('div'), {classes: 'login-box-close', parent: loginBox});

        this.root.classes('root-min-height');

        this.onClick(forgotLink,  () => this.forgotPassword());
        this.onClick(loginLink,   () => this.popupSwitch());
        this.onClick(close,       () => this.loginBoxClose());

        let hasOAuth = false;
        const oauthProviders = ['google', 'github', 'gitlab'];
        oauthProviders.filter(p => this.configuredOauths[p]).forEach(provider => {
            const button = Wrap.new(
                'b)utton',
                {classes: ['button', `${provider}-button`], type: 'button', innerText: provider, parent: oauthButtons});
            this.onClick(button, () => this.commentoAuth(provider, commentHex));
            hasOAuth = true;
        });

        if (this.configuredOauths['sso']) {
            const button = Wrap.new(
                'b)utton',
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
        const emailSubtitle = Wrap.byId(IDS.loginBoxEmailSubtitle);

        if (this.oauthButtonsShown) {
            this.remove(
                Wrap.byId(IDS.loginBoxOauthButtonsContainer),
                Wrap.byId(IDS.loginBoxOauthPretext),
                Wrap.byId(IDS.loginBoxHr1),
                Wrap.byId(IDS.loginBoxHr2));
        }

        if (this.configuredOauths['sso']) {
            this.remove(
                Wrap.byId(IDS.loginBoxSsoButtonContainer),
                Wrap.byId(IDS.loginBoxSsoPretext),
                Wrap.byId(IDS.loginBoxHr1),
                Wrap.byId(IDS.loginBoxHr2));
        }

        this.remove(Wrap.byId(IDS.loginBoxLoginLinkContainer), Wrap.byId(IDS.loginBoxForgotLinkContainer));

        emailSubtitle.innerText = 'Create an account';
        this.popupBoxType = 'signup';
        this.showPasswordField();
        Wrap.byId(IDS.loginBoxEmailInput).focus();
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
                this.remove(Wrap.byId(IDS.login));
                return (id ? this.commentNew(id, resp.commenterToken, false) : undefined);
            })
            .then(() => this.loginBoxClose())
            .then(() => this.commentsGet())
            .then(() => this.commentsRender())
            .then(() => this.allShow());
    }

    login(id: string): Promise<void> {
        const email    = Wrap.byId(IDS.loginBoxEmailInput);
        const password = Wrap.byId(IDS.loginBoxPasswordInput);
        return this.loginUP(email.value, password.value, id);
    }

    signup(id: string): Promise<void> {
        const email    = Wrap.byId(IDS.loginBoxEmailInput);
        const name     = Wrap.byId(IDS.loginBoxNameInput);
        const website  = Wrap.byId(IDS.loginBoxWebsiteInput);
        const password = Wrap.byId(IDS.loginBoxPasswordInput);

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
        const loginBox = Wrap.byId(IDS.loginBox);
        const subtitle = Wrap.byId(IDS.loginBoxEmailSubtitle);

        this.remove(
            Wrap.byId(IDS.loginBoxEmailButton),
            Wrap.byId(IDS.loginBoxLoginLinkContainer),
            Wrap.byId(IDS.loginBoxForgotLinkContainer));
        if (this.oauthButtonsShown && Object.keys(this.configuredOauths).length) {
            this.remove(
                Wrap.byId(IDS.loginBoxHr1),
                Wrap.byId(IDS.loginBoxHr2),
                Wrap.byId(IDS.loginBoxOauthPretext),
                Wrap.byId(IDS.loginBoxOauthButtonsContainer));
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
            const fieldContainer = Wrap.new('div'), {classes: 'email-container'});
            const field          = Wrap.new('div'), {classes: 'email', parent: fieldContainer});
            const fieldInput     = Wrap.new('input'), c);
            this.append(field, fieldInput);
            // Add a submit button next to the password input
            if (c.type === 'password') {
                Wrap.new('button'), {
                    id:        IDS.loginBoxPasswordButton,
                    type:      'submit',
                    classes:   'email-button',
                    innerText: this.popupBoxType,
                    parent:    field,
                });
            }
            this.append(loginBox, fieldContainer);
        });

        Wrap.byId(isSignup ? IDS.loginBoxNameInput : IDS.loginBoxPasswordInput).focus();
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
        const lock = Wrap.byId(IDS.modToolsLockButton);
        this.isLocked = !this.isLocked;
        lock.disabled = true;
        return this.pageUpdate()
            .then(() => lock.disabled = false)
            .then(() => this.reload());
    }

    commentSticky(commentHex: string): Promise<void> {
        if (this.stickyCommentHex !== 'none') {
            const sticky = Wrap.byId(IDS.sticky + this.stickyCommentHex);
            this.removeClasses(sticky, 'option-unsticky');
            this.addClasses(sticky, 'option-sticky');
        }

        this.stickyCommentHex = this.stickyCommentHex === commentHex ? 'none' : commentHex;

        return this.pageUpdate()
            .then(() => {
                const sticky = Wrap.byId(IDS.sticky + commentHex);
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
        Wrap.new('div').id(IDS.mainArea).classes('main-area').style('display: none').appendTo(this.root);
    }

    modToolsCreate() {
        Wrap.new('div').id(IDS.modTools).classes('mod-tools').style('display: none').appendTo(this.root)
            .append(
                Wrap.new('button')
                    .id(IDS.modToolsLockButton)
                    .attr({type: 'button'})
                    .inner(this.isLocked ? 'Unlock Thread' : 'Lock Thread')
                    .click(() => this.threadLockToggle()));
    }

    allShow() {
        const mainArea = Wrap.byId(IDS.mainArea);
        const modTools = Wrap.byId(IDS.modTools);
        const loggedContainer = Wrap.byId(IDS.loggedContainer);

        this.setAttr(mainArea, {style: null});

        if (this.isModerator) {
            this.setAttr(modTools, {style: null});
        }

        if (loggedContainer) {
            this.setAttr(loggedContainer, {style: null});
        }
    }

    loginBoxClose() {
        const mainArea = Wrap.byId(IDS.mainArea);
        const loginBoxContainer = Wrap.byId(IDS.loginBoxContainer);

        this.removeClasses(mainArea, 'blurred');
        this.root.noClasses('root-min-height');

        this.setAttr(loginBoxContainer, {style: 'display: none'});
    }

    loginBoxShow(commentHex: string) {
        const mainArea = Wrap.byId(IDS.mainArea);
        const loginBoxContainer = Wrap.byId(IDS.loginBoxContainer);

        this.popupRender(commentHex);

        this.addClasses(mainArea, 'blurred');
        this.setAttr(loginBoxContainer, {style: null});

        loginBoxContainer.scrollIntoView({behavior: 'smooth'});

        Wrap.byId(IDS.loginBoxEmailInput).focus();
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
                const el = Wrap.byId(IDS.card + id);
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
                this.root.scrollTo();
            }
        }
    }
}
