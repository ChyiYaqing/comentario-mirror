import { HttpClient } from './http-client';
import {
    Comment,
    Commenter,
    CommenterMap,
    CommentMap,
    CommentsGroupedByHex,
    Email,
    sortingProps,
    SortPolicy,
    StringBooleanMap,
} from './models';
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
import { LoginDialog } from './login-dialog';
import { SignupDialog } from './signup-dialog';
import { UIToolkit } from './ui-toolkit';
import { MarkdownHelp } from './markdown-help';
import { CommentCard, CommentRenderingContext, CommentTree } from './comment-card';
import { Utils } from './utils';

const IDS = {
    loginBtn:          'login-btn',
    superContainer:    'textarea-super-container-',
    textarea:          'textarea-',
    anonymousCheckbox: 'anonymous-checkbox-',
    sortPolicy:        'sort-policy-',
    card:              'comment-card-',
    text:              'comment-text-',
    edit:              'comment-edit-',
    reply:             'comment-reply-',
};

export class Comentario {

    /** Origin URL, which gets replaced by the backend on serving the file. */
    private readonly origin = '[[[.Origin]]]';
    /** CDN URL, which gets replaced by the backend on serving the file. */
    private readonly cdn = '[[[.CdnPrefix]]]';
    /** App version, which gets replaced by the backend on serving the file. */
    private readonly version = '[[[.Version]]]';

    /** HTTP client we'll use for API requests. */
    private readonly apiClient = new HttpClient(`${this.origin}/api`);

    /** Default ID of the container element Comentario will be embedded into. */
    private rootId = 'comentario';

    /** The root element of Comentario embed. */
    private root: Wrap<any>;

    /** Error message panel (only shown when needed). */
    private error: Wrap<HTMLDivElement>;

    /** Moderator tools panel. */
    private modTools: Wrap<HTMLDivElement>;
    private modToolsLockBtn: Wrap<HTMLButtonElement>;

    /** Main area panel. */
    private mainArea: Wrap<HTMLDivElement>;

    /** Comments panel inside the mainArea. */
    private commentsArea: Wrap<HTMLDivElement>;

    private pageId = parent.location.pathname;
    private cssOverride: string;
    private noFonts = false;
    private hideDeleted = false;
    private autoInit = true;
    private isAuthenticated = false;
    private comments: Comment[] = [];

    /** Loaded comment objects indexed by commentHex. */
    private commentsByHex: CommentMap = {};

    /** Map of commenters by their hsx ID. */
    private readonly commenters: CommenterMap = {};
    private requireIdentification = true;
    private isModerator = false;
    private isFrozen = false;
    private chosenAnonymous = false;
    private isLocked = false;
    private stickyCommentHex = 'none';
    private shownReply: StringBooleanMap;
    private readonly shownEdit: StringBooleanMap = {};
    private configuredOauths: StringBooleanMap = {};
    private anonymousOnly = false;
    private sortPolicy: SortPolicy = 'score-desc';
    private selfHex: string = undefined;
    private readonly loadedCss: StringBooleanMap = {};
    private initialised = false;

    constructor(
        private readonly doc: Document,
    ) {
        this.whenDocReady().then(() => this.init());
    }

    /**
     * Retrieve a token of the authenticated user. If the user isn't authenticated, return 'anonymous'.
     */
    get token(): string {
        return `; ${this.doc.cookie}`.split('; comentario_auth_token=').pop().split(';').shift() || 'anonymous';
    }

    /**
     * Store a token of the authenticated user in a cookie.
     */
    set token(v: string) {
        // Set the cookie expiration date one year in the future
        const date = new Date();
        date.setTime(date.getTime() + (365 * 24 * 60 * 60 * 1000));

        // Store the cookie
        this.doc.cookie = `comentario_auth_token=${v}; expires=${date.toUTCString()}; path=/`;
    }

    /**
     * The main worker routine of Comentario
     * @return Promise that resolves as soon as Comentario setup is complete
     */
    async main(): Promise<void> {
        // Make sure there's a root element present, and save it
        this.root = Wrap.byId(this.rootId, true);
        if (!this.root.ok) {
            return this.reject(`No root element with id='${this.rootId}' found. Check your configuration and HTML.`);
        }

        this.root.classes('root', !this.noFonts && 'root-font');

        // Begin by loading the stylesheet
        await this.cssLoad(`${this.cdn}/css/comentario.css`);

        // Load stylesheet override, if any
        if (this.cssOverride) {
            await this.cssLoad(this.cssOverride);
        }

        // Load the UI
        await this.reload();

        // Scroll to the requested comment, if any
        this.scrollToCommentHash();
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

    /**
     * Initialise the Comentario engine on the current page.
     * @private
     */
    private async init(): Promise<void> {
        // Only perform initialisation once
        if (this.initialised) {
            return this.reject('Already initialised, ignoring the repeated init call');
        }
        this.initialised = true;

        // Parse any custom data-* tags on the Comentario script element
        this.dataTagsLoad();

        // If automatic initialisation is activated (default), run Comentario
        if (this.autoInit) {
            await this.main();
        }
        console.info(`Initialised Comentario ${this.version}`);
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
                        Wrap.new('link').attr({href: url, rel: 'stylesheet', type: 'text/css'}).on('load', () => resolve()));
            });
    }

    /**
     * Create a new editor for editing comment text.
     * @param commentHex Comment's hex ID.
     * @param isEdit Whether it's adding a new comment (false) or editing an existing one (true)
     */
    textareaCreate(commentHex: string, isEdit: boolean): Wrap<HTMLFormElement> {
        // "Comment anonymously" checkbox
        let anonContainer: Wrap<any>;
        if (!this.requireIdentification && !isEdit) {
            const anonCheckbox = Wrap.new('input').id(IDS.anonymousCheckbox + commentHex).attr({type: 'checkbox'});
            if (this.anonymousOnly) {
                anonCheckbox.checked(true).attr({disabled: 'true'});
            }
            anonContainer = UIToolkit.div('round-check', 'anonymous-checkbox-container')
                .append(
                    anonCheckbox,
                    Wrap.new('label').attr({for: Wrap.idPrefix + IDS.anonymousCheckbox + commentHex}).inner('Comment anonymously'));
        }

        // Instantiate and set up a new form
        return UIToolkit.form(() => isEdit ? this.submitCommentEdits(commentHex) : this.submitNewComment(commentHex))
            .id(IDS.superContainer + commentHex)
            .classes('textarea-form')
            .append(
                // Textarea in a container
                UIToolkit.div('textarea-container')
                    .append(UIToolkit.textarea('Add a comment', true, true).id(IDS.textarea + commentHex)),
                // Textarea footer
                UIToolkit.div('textarea-form-footer')
                    .append(
                        UIToolkit.div()
                            .append(
                                // Anonymous checkbox, if any
                                anonContainer,
                                // Markdown help button
                                UIToolkit.button(
                                    '<b>M⬇</b>&nbsp;Markdown',
                                    btn => MarkdownHelp.run(this.root, {ref: btn, placement: 'bottom-start'}))),
                        // Submit button
                        UIToolkit.submit(isEdit ? 'Save Changes' : 'Add Comment', false)));
    }

    sortPolicyApply(policy: SortPolicy) {
        Wrap.byId(IDS.sortPolicy + this.sortPolicy).noClasses('sort-policy-button-selected');
        Wrap.byId(IDS.sortPolicy + policy).classes('sort-policy-button-selected');
        this.sortPolicy = policy;

        // Re-render the sorted comment
        this.renderComments();
    }

    /**
     * Create and return a toolbar with sort policy buttons.
     * @private
     */
    private sortPolicyBar(): Wrap<HTMLDivElement> {
        return UIToolkit.div('sort-policy-buttons-container')
            .append(
                UIToolkit.div('sort-policy-buttons')
                    .append(
                        ...Object.keys(sortingProps).map((sp: SortPolicy) =>
                            Wrap.new('a')
                                .id(IDS.sortPolicy + sp)
                                .classes('sort-policy-button', sp === this.sortPolicy && 'sort-policy-button-selected')
                                .inner(sortingProps[sp].label)
                                .click(() => this.sortPolicyApply(sp)))));
    }

    /**
     * Create a new editor for editing a comment with the given hex ID.
     * @param commentHex Comment's hex ID.
     */
    startEditing(commentHex: string) {
        if (this.shownEdit[commentHex]) {
            return;
        }

        this.shownEdit[commentHex] = true;
        Wrap.byId(IDS.text + commentHex).replaceWith(this.textareaCreate(commentHex, true));
        Wrap.byId(IDS.textarea + commentHex).value(this.commentsByHex[commentHex].markdown);

        // Turn the Edit button into a Cancel edit button
        Wrap.byId(IDS.edit + commentHex)
            .noClasses('option-edit')
            .classes('option-cancel')
            .attr({title: 'Cancel edit'})
            .unlisten()
            .click(() => this.stopEditing(commentHex));
    }

    /**
     * Close the created editor for editing a comment with the given hex ID, cancelling the edits.
     * @param commentHex Comment's hex ID.
     */
    stopEditing(commentHex: string) {
        delete this.shownEdit[commentHex];
        Wrap.byId(IDS.superContainer + commentHex)
            .html(this.commentsByHex[commentHex].html)
            .id(IDS.text + commentHex);

        // Turn the Cancel edit button back into the Edit button
        Wrap.byId(IDS.edit + commentHex)
            .noClasses('option-cancel')
            .classes('option-edit')
            .attr({title: 'Edit comment'})
            .unlisten()
            .click(() => this.startEditing(commentHex));
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

        this.shownReply[commentHex] = true;
        this.textareaCreate(commentHex, false).insertAfter(Wrap.byId(IDS.text + commentHex));
        Wrap.byId(IDS.reply + commentHex)
            .noClasses('option-reply')
            .classes('option-cancel')
            .attr({title: 'Cancel reply'})
            .unlisten()
            .click(() => this.replyCollapse(commentHex));
    }

    /**
     * Close the created editor for editing a reply to the comment with the given hex ID.
     * @param commentHex Comment's hex ID.
     */
    replyCollapse(commentHex: string) {
        delete this.shownReply[commentHex];
        Wrap.byId(IDS.superContainer + commentHex).remove();
        Wrap.byId(IDS.reply + commentHex)
            .noClasses('option-cancel')
            .classes('option-reply')
            .attr({title: 'Reply to this comment'})
            .unlisten()
            .click(() => this.replyShow(commentHex));
    }

    dataTagsLoad() {
        for (const script of this.doc.getElementsByTagName('script')) {
            if (script.src.match(/\/js\/comentario\.js$/)) {
                const ws = new Wrap(script);
                let s = ws.getAttr('data-page-id');
                if (s) {
                    this.pageId = s;
                }
                this.cssOverride = ws.getAttr('data-css-override');
                this.autoInit = ws.getAttr('data-auto-init') !== 'false';
                s = ws.getAttr('data-id-root');
                if (s) {
                    this.rootId = s;
                }
                this.noFonts = ws.getAttr('data-no-fonts') === 'true';
                this.hideDeleted = ws.getAttr('data-hide-deleted') === 'true';
                break;
            }
        }
    }

    /**
     * Scroll to the comment whose hex ID is provided in the current window's fragment (if any).
     * @private
     */
    private scrollToCommentHash() {
        const h = window.location.hash;

        // If the hash starts with a valid hex ID
        if (h?.startsWith('#comentario-')) {
            const id = h.substring(10);
            Wrap.byId(IDS.card + id)
                .classes('highlighted-card')
                .scrollTo()
                .else(() => {
                    // Make sure it's a (sort of) valid ID before showing the user a message
                    if (id.length === 64) {
                        this.setError('The comment you\'re looking for doesn\'t exist; possibly it was deleted.');
                    }
                });


        } else if (h?.startsWith('#comentario')) {
            // If we're requested to scroll to the comments in general
            this.root.scrollTo();
        }
    }

    /**
     * (Re)render all comments recursively, adding them to the comments area.
     * @private
     */
    private renderComments() {
        this.commentsArea
            .html('')
            .append(new CommentTree().render(this.makeCommentRenderingContext(), 'root'));
    }

    /**
     * Set and display (message is given) or clean (message is falsy) an error message in the error panel.
     * @param message Message to set. If falsy, the error panel gets removed.
     * @return Whether there was a (truthy) error.
     * @private
     */
    private setError(message?: string): boolean {
        if (message) {
            this.error = (this.error || UIToolkit.div('error-box').prependTo(this.root)).inner(message);
            return true;
        }
        this.error?.remove();
        this.error = undefined;
        return false;
    }

    /**
     * Request the authentication status of the current user from the backend, and return a promise that resolves as
     * soon as the status becomes definite.
     * @private
     */
    private async getAuthStatus(): Promise<void> {
        this.isAuthenticated = false;

        // If we're already (knowingly) anonymous
        const token = this.token;
        if (token !== 'anonymous') {
            // Fetch the status from the backend
            try {
                const r = await this.apiClient.post<ApiSelfResponse>('commenter/self', {commenterToken: token});
                if (!r.success) {
                    this.token = 'anonymous';
                } else {
                    this.setupCurUserProfile(r.commenter, r.email);
                    this.isAuthenticated = true;
                }
            } catch (e) {
                // On any error consider the user unauthenticated
                console.error(e);
            }
        }
    }

    /**
     * Reload the app UI.
     */
    private async reload() {
        // Remove any content from the root
        this.root.html('');
        this.modTools = null;
        this.modToolsLockBtn = null;
        this.mainArea = null;
        this.commentsArea = null;
        this.shownReply = {};

        // Load information about ourselves
        await this.getAuthStatus();

        // Fetch page data and comments
        await this.loadPageData();

        // Create the layout
        this.root.append(
            // Moderator toolbar
            this.isModerator && this.createModToolsPanel(),
            // Main area
            this.createMainArea(),
            // Footer
            this.createFooter(),
        );

        // Render the comments
        this.renderComments();
    }

    /**
     * Create and return a moderator toolbar element.
     * @private
     */
    private createModToolsPanel(): Wrap<HTMLDivElement> {
        this.modToolsLockBtn = UIToolkit.button(
            this.isLocked ? 'Unlock thread' : 'Lock thread',
            () => this.threadLockToggle());
        this.modTools = UIToolkit.div('mod-tools')
            .append(Wrap.new('span').classes('mod-tools-title').inner('Moderator tools'), this.modToolsLockBtn)
            .appendTo(this.root);
        return this.modTools;
    }

    /**
     * Create and return a main area element.
     * @private
     */
    private createMainArea(): Wrap<HTMLDivElement> {
        this.mainArea = UIToolkit.div('main-area');

        // If there's any auth provider configured
        if (Object.values(this.configuredOauths).some(b => b)) {
            // If not authenticated, add a Login button
            if (!this.isAuthenticated) {
                UIToolkit.div('login')
                    .append(UIToolkit.button('Login', () => this.showLoginDialog(null), 'fw-bold').id(IDS.loginBtn))
                    .appendTo(this.mainArea);
            }

        } else if (!this.requireIdentification) {
            // No auth provider available, but we allow anonymous commenting
            this.anonymousOnly = true;
        }

        // If commenting is locked/frozen, add a corresponding message
        if (this.isLocked || this.isFrozen) {
            if (this.isAuthenticated || this.chosenAnonymous) {
                this.mainArea.append(UIToolkit.div('moderation-notice').inner('This thread is locked. You cannot add new comments.'));
            }

        // Otherwise, add a root editor (for creating a new comment)
        } else {
            this.mainArea.append(this.textareaCreate('root', false));
        }

        // If there's any comment, add sort buttons
        if (this.comments.length) {
            this.mainArea.append(this.sortPolicyBar());
        }

        // Create a panel for comments
        this.commentsArea = UIToolkit.div('comments').appendTo(this.mainArea);
        return this.mainArea;
    }

    /**
     * Create and return a footer panel.
     * @private
     */
    private createFooter(): Wrap<HTMLDivElement> {
        return UIToolkit.div('footer')
            .append(
                UIToolkit.div('logo-container')
                    .append(
                        Wrap.new('a')
                            .attr({href: 'https://comentario.app/', target: '_blank'})
                            .html('Powered by ')
                            .append(Wrap.new('span').classes('logo-brand').inner('Comentario'))));
    }

    /**
     * Only called when there's an authenticated user. Sets up the controls related to the current user.
     * @param commenter Currently authenticated user.
     * @param email Email of the commenter.
     * @private
     */
    private setupCurUserProfile(commenter: Commenter, email: Email) {
        this.commenters[commenter.commenterHex] = commenter;
        this.selfHex = commenter.commenterHex;

        // Create an avatar element
        const idxColor = Utils.colourIndex(`${commenter.commenterHex}-${commenter.name}`);
        const avatar = commenter.photo === 'undefined' ?
            UIToolkit.div('avatar', `bg-${idxColor}`).html(commenter.name[0].toUpperCase()) :
            Wrap.new('img')
                .classes('avatar-img')
                .attr({src: `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`, loading: 'lazy', alt: ''});

        // Create a profile bar
        const link = !commenter.link || commenter.link === 'undefined' ? undefined : commenter.link;
        UIToolkit.div('profile-bar')
            .append(
                // Commenter avatar and name
                UIToolkit.div('logged-in-as')
                    .append(
                        // Avatar
                        avatar,
                        // Name and link
                        Wrap.new(link ? 'a' : 'div')
                            .classes('name')
                            .inner(commenter.name)
                            .attr({href: link, rel: link && 'nofollow noopener noreferrer'})),
                // Buttons on the right
                UIToolkit.div()
                    .append(
                        // If it's a local user, add a Profile link
                        commenter.provider === 'commento' &&
                            Wrap.new('a')
                                .classes('profile-link')
                                .inner('Profile')
                                .attr({href: `${this.origin}/profile?commenterToken=${this.token}`, target: '_blank'}),
                        // Notifications link
                        Wrap.new('a')
                            .classes('profile-link')
                            .inner('Notifications')
                            .attr({href: `${this.origin}/unsubscribe?unsubscribeSecretHex=${email.unsubscribeSecretHex}`, target: '_blank'}),
                        // Logout link
                        Wrap.new('a')
                            .classes('profile-link')
                            .inner('Logout')
                            .attr({href: ''})
                            .click((_, e) => this.logout(e))))
            .prependTo(this.root);
    }

    /**
     * Register the user with the given details and log them in.
     * @param name User's full name.
     * @param website User's website.
     * @param email User's email.
     * @param password User's password.
     * @param commentHex Optional comment hex ID to add.
     */
    private async signup(name: string, website: string, email: string, password: string, commentHex: string): Promise<void> {
        // Sign the user up
        const r = await this.apiClient.post<ApiResponseBase>('commenter/new', {name, website, email, password});
        if (this.setError(!r.success && r.message)) {
            return Promise.reject();
        }

        // Log the user in, submitting their comment (if any)
        return this.authenticateLocally(email, password, commentHex);
    }

    /**
     * Authenticate the user using local authentication (email and password).
     * @param email User's email.
     * @param password User's password.
     * @param commentHex Optional comment hex ID to add.
     */
    private async authenticateLocally(email: string, password: string, commentHex: string): Promise<void> {
        // Log the user in
        const r = await this.apiClient.post<ApiCommenterLoginResponse>('commenter/login', {email, password});
        if (this.setError(!r.success && r.message)) {
            return Promise.reject();
        }

        // Store the authenticated token in a cookie
        this.token = r.commenterToken;

        // Submit a new comment, if needed
        if (commentHex) {
            await this.commentNew(commentHex, r.commenterToken, false);
        }

        // Reload the whole bunch
        return this.reload();
    }

    /**
     * Show the signup dialog.
     * @param commentHex Optional comment hex ID to add upon signup.
     * @private
     */
    private async showSignupDialog(commentHex: string): Promise<void> {
        const dlg = await SignupDialog.run(
            this.root,
            {ref: Wrap.byId(IDS.loginBtn), placement: 'bottom-end'});
        return dlg.confirmed && await this.signup(dlg.name, dlg.website, dlg.email, dlg.password, commentHex);
    }

    /**
     * Show the login dialog.
     * @param commentHex Optional comment hex ID to add upon login.
     * @private
     */
    private async showLoginDialog(commentHex: string): Promise<void> {
        const dlg = await LoginDialog.run(
            this.root,
            {ref: Wrap.byId(IDS.loginBtn), placement: 'bottom-end'},
            this.configuredOauths,
            this.origin);
        if (dlg.confirmed) {
            switch (dlg.navigateTo) {
                case null:
                    // Local auth
                    return await this.authenticateLocally(dlg.email, dlg.password, commentHex);

                case 'forgot':
                    // Already navigated to the Forgot password page in a new tab
                    return;

                case 'signup':
                    // Switch to signup
                    return await this.showSignupDialog(commentHex);

                default:
                    // External auth
                    return await this.openOAuthPopup(dlg.navigateTo, commentHex);
            }
        }
    }

    /**
     * Open a new browser popup window for authenticating with the given identity provider.
     * @param idp Identity provider to initiate authentication with.
     * @param commentHex Optional hex ID of the comment to add upon successful authentication.
     * @private
     */
    private async openOAuthPopup(idp: string, commentHex: string): Promise<void> {
        // Request a token
        const r = await this.apiClient.get<ApiCommenterTokenNewResponse>('commenter/token/new');
        if (this.setError(!r.success && r.message)) {
            return this.reject(r.message);
        }

        // Store the obtained auth token
        this.token = r.commenterToken;

        // Open a popup window
        const popup = window.open(
            `${this.origin}/api/oauth/${idp}/redirect?commenterToken=${r.commenterToken}`,
            '_blank',
            'popup,width=800,height=600');

        // Wait until the popup is closed
        await new Promise<void>(resolve => {
            const interval = setInterval(
                () => {
                    if (popup.closed) {
                        clearInterval(interval);
                        resolve();
                    }
                },
                500);
        });

        // Refresh the auth status
        await this.getAuthStatus();

        // Submit the pending comment, if there was one
        if (this.isAuthenticated && commentHex) {
            await this.commentNew(commentHex, this.token, false);
        }

        // Reload the whole bunch
        return this.reload();
    }

    /**
     * Log the current user out.
     * @param e Click event that triggered the logout.
     * @private
     */
    private logout(e: MouseEvent): Promise<void> {
        e.preventDefault();
        this.token = 'anonymous';
        this.isAuthenticated = false;
        this.isModerator = false;
        this.selfHex = undefined;
        return this.reload();
    }

    /**
     * Load data for the current page URL, including the comments, from the backend and store them locally
     * @private
     */
    private async loadPageData(): Promise<void> {
        // Retrieve a comment list from the backend
        const r = await this.apiClient.post<ApiCommentListResponse>('comment/list', {
            commenterToken: this.token,
            domain:         parent.location.host,
            path:           this.pageId,
        });
        if (this.setError(!r.success && r.message)) {
            return;
        }

        // Store all known commenters
        Object.assign(this.commenters, r.commenters);

        // Store page- and backend-related properties
        this.requireIdentification = r.requireIdentification;
        this.isModerator           = r.isModerator;
        this.isFrozen              = r.isFrozen;
        this.isLocked              = r.attributes.isLocked;
        this.stickyCommentHex      = r.attributes.stickyCommentHex;
        this.configuredOauths      = r.configuredOauths;
        this.sortPolicy            = r.defaultSortPolicy;

        // Update comment models and make a hex-comment map
        this.comments = r.comments;
        this.commentsByHex = {};
        this.comments.forEach(c => {
            c.creationMs = new Date(c.creationDate).getTime();
            this.commentsByHex[c.commentHex] = c;
        });
    }

    /**
     * Submit a new comment with the given hex ID, forcing the user to authenticate, if needed.
     * @param commentHex Comment's hex ID.
     * @private
     */
    private async submitNewComment(commentHex: string): Promise<void> {
        if (this.requireIdentification || !Wrap.byId(IDS.anonymousCheckbox + commentHex).isChecked) {
            return this.isAuthenticated ?
                this.commentNew(commentHex, this.token, true) :
                this.showLoginDialog(commentHex);
        }

        this.chosenAnonymous = true;
        return this.commentNew(commentHex, 'anonymous', true);
    }

    /**
     * Submit the entered comment markdown to the backend for saving.
     * @param commentHex Comment's hex ID
     */
    private async submitCommentEdits(commentHex: string): Promise<void> {
        const textarea = Wrap.byId(IDS.textarea + commentHex);

        // Validate the textarea value
        if (!textarea.valid) {
            return Promise.reject();
        }

        // Submit the edit to the backend
        const markdown = textarea.val.trim();
        const r = await this.apiClient.post<ApiCommentEditResponse>('comment/edit', {commenterToken: this.token, commentHex, markdown});
        if (this.setError(!r.success && r.message)) {
            return;
        }

        // Update the locally stored comment's data
        this.commentsByHex[commentHex].markdown = markdown;
        this.commentsByHex[commentHex].html = r.html;

        // Hide the editor
        this.stopEditing(commentHex);

        // Update the comment's moderation notice
        this.updateCommentModerationNotice(commentHex, r.state);
    }

    /**
     * Submit a new comment entered under the given hex ID.
     * @param commentHex Comment's hex ID.
     * @param commenterToken Token of the current commenter.
     * @param appendCard Whether to also add a new card for the created comment.
     * @private
     */
    private async commentNew(commentHex: string, commenterToken: string, appendCard: boolean): Promise<void> {
        // Validate the textarea value
        const textarea  = Wrap.byId(IDS.textarea + commentHex);
        if (!textarea.valid) {
            return Promise.reject();
        }

        // Submit the comment to the backend
        const markdown = textarea.val.trim();
        const r = await this.apiClient.post<ApiCommentNewResponse>('comment/new', {
            commenterToken,
            domain:    parent.location.host,
            path:      this.pageId,
            parentHex: commentHex,
            markdown,
        });
        if (this.setError(!r.success && r.message)) {
            return;
        }

        // Update the comment's moderation notice
        this.updateCommentModerationNotice(commentHex, r.state);

        // Store the updated comment in the local map
        const comment: Comment = {
            commentHex:   r.commentHex,
            commenterHex: this.selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : this.selfHex,
            markdown,
            html:         r.html,
            parentHex:    'root',
            score:        0,
            state:        'approved',
            direction:    0,
            creationDate: new Date().toISOString(),
            deleted:      false,
        };
        this.commentsByHex[r.commentHex] = comment;

        // Remove the entered comment text and reset its touched state
        textarea.value('').noClasses('touched');

        // Add the new card, if needed
        if (appendCard) {
            const newCard = new CommentTree().render(this.makeCommentRenderingContext({root: [comment]}), 'root');
            if (commentHex === 'root') {
                newCard.prependTo(this.commentsArea);
            } else {
                Wrap.byId(IDS.superContainer + commentHex).replaceWith(newCard);
                this.shownReply[commentHex] = false;
                Wrap.byId(IDS.reply + commentHex)
                    .noClasses('option-cancel')
                    .classes('option-reply')
                    .attr({title: 'Reply to this comment'})
                    .click(() => this.replyShow(commentHex));
            }
        }
    }

    /**
     * Add the relevant moderation notice to the given comment, if needed.
     * @param commentHex Comment's hex ID.
     * @param state Comment's moderation state.
     * @private
     */
    private updateCommentModerationNotice(commentHex: string, state: 'unapproved' | 'flagged') {
        let message = '';
        switch (state) {
            case 'unapproved':
                message = 'Your comment is under moderation.';
                break;
            case 'flagged':
                message = 'Your comment was flagged as spam and is under moderation.';
                break;
            default:
                return;
        }
        UIToolkit.div('moderation-notice').inner(message).prependTo(Wrap.byId(IDS.superContainer + commentHex));
    }

    /**
     * Toggle the current comment's thread lock status.
     * @private
     */
    private async threadLockToggle(): Promise<void> {
        this.modToolsLockBtn.attr({disabled: 'true'});
        this.isLocked = !this.isLocked;
        await this.submitPageAttrs();
        this.modToolsLockBtn.attr({disabled: 'false'});
        return this.reload();
    }

    /**
     * Approve the comment of the given card.
     * @private
     */
    private async commentApprove(card: CommentCard): Promise<void> {
        // Submit the approval to the backend
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/approve',
            {commenterToken: this.token, commentHex: card.comment.commentHex});
        if (this.setError(!r.success && r.message)) {
            return;
        }

        // Update the comment and card
        card.comment.state = 'approved';
        card.update();
    }

    /**
     * Delete the comment of the given card.
     * @private
     */
    private async commentDelete(card: CommentCard): Promise<void> {
        // Run deletion with the backend
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/delete',
            {commenterToken: this.token, commentHex: card.comment.commentHex});
        if (this.setError(!r.success && r.message)) {
            return;
        }

        // Update the comment and card
        card.comment.deleted = true;
        card.update();
    }

    /**
     * Toggle the given comment's sticky status.
     * @private
     */
    private async commentSticky(card: CommentCard): Promise<void> {
        // Save the page's sticky comment ID
        this.stickyCommentHex = this.stickyCommentHex === card.comment.commentHex ? 'none' : card.comment.commentHex;
        await this.submitPageAttrs();

        // Reload the comments
        return this.reload();
    }

    /**
     * Vote (upvote, downvote, or undo vote) for the given comment.
     * @private
     */
    private async commentVote(card: CommentCard, direction: -1 | 0 | 1): Promise<void> {
        // Only registered users can vote
        if (!this.isAuthenticated) {
            return this.showLoginDialog(null);
        }

        // Run the vote with the API
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/vote',
            {commenterToken: this.token, commentHex: card.comment.commentHex, direction});
        if (this.setError(!r.success && r.message)) {
            return Promise.reject();
        }

        // Update the vote and the score
        card.comment.score += direction - card.comment.direction;
        card.comment.direction = direction;

        // Update the card
        card.update();
    }

    /**
     * Submit the currently set page state (sticky comment and lock) to the backend.
     * @private
     */
    private async submitPageAttrs(): Promise<void> {
        const r = await this.apiClient.post<ApiResponseBase>('page/update', {
            commenterToken: this.token,
            domain:         parent.location.host,
            path:           this.pageId,
            attributes:     {isLocked: this.isLocked, stickyCommentHex: this.stickyCommentHex},
        });
        this.setError(!r.success && r.message);
    }

    /**
     * Return a new comment rendering context.
     * @param parentMap Optional parent map to use. If not provided, a new one is created based on all available comments.
     */
    private makeCommentRenderingContext(parentMap?: CommentsGroupedByHex): CommentRenderingContext {
        // If no parent map provided, group comments by parent hex ID: make map {parentHex: Comment[]}
        if (!parentMap) {
            parentMap = this.comments.reduce(
                (m, c) => {
                    const ph = c.parentHex;
                    if (ph in m) {
                        m[ph].push(c);
                    } else {
                        m[ph] = [c];
                    }
                    return m;
                },
                {} as CommentsGroupedByHex);
        }

        // Make a new context instance
        return {
            cdn:             this.cdn,
            root:            this.root,
            parentMap,
            commenters:      this.commenters,
            selfHex:         this.selfHex,
            stickyHex:       this.stickyCommentHex,
            sortPolicy:      this.sortPolicy,
            isAuthenticated: this.isAuthenticated,
            isModerator:     this.isModerator,
            hideDeleted:     this.hideDeleted,
            curTimeMs:       new Date().getTime(),
            onApprove:       card => this.commentApprove(card),
            onDelete:        card => this.commentDelete(card),
            onEdit:          () => { /*TODO*/ },
            onReply:         () => { /*TODO*/ },
            onSticky:        card => this.commentSticky(card),
            onVote:          (card, direction) => this.commentVote(card, direction),
        };
    }
}
