import { HttpClient } from './http-client';
import { Comment, CommenterMap, CommentsGroupedByHex, SortPolicy, StringBooleanMap } from './models';
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
import { UIToolkit } from './ui-toolkit';
import { CommentCard, CommentRenderingContext, CommentTree } from './comment-card';
import { CommentEditor } from './comment-editor';
import { ProfileBar } from './profile-bar';
import { SortBar } from './sort-bar';

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
    private error?: Wrap<HTMLDivElement>;

    /** User profile toolbar. */
    private profileBar: ProfileBar;

    /** Moderator tools panel. */
    private modTools: Wrap<HTMLDivElement>;
    private modToolsLockBtn: Wrap<HTMLButtonElement>;

    /** Main area panel. */
    private mainArea: Wrap<HTMLDivElement>;

    /** Container for hosting the Add comment editor. */
    private addCommentHost: Wrap<HTMLDivElement>;

    /** Currently active comment editor instance. */
    private editor?: CommentEditor;

    /** Comments panel inside the mainArea. */
    private commentsArea: Wrap<HTMLDivElement>;

    /** Map of commenters by their hsx ID. */
    private readonly commenters: CommenterMap = {};

    /** Map of loaded CSS stylesheet URLs. */
    private readonly loadedCss: StringBooleanMap = {};

    /** Map of comments, grouped by their parentHex. */
    private parentHexMap?: CommentsGroupedByHex;

    private pageId = parent.location.pathname;
    private cssOverride: string;
    private noFonts = false;
    private hideDeleted = false;
    private autoInit = true;
    private requireIdentification = true;
    private isAuthenticated = false;
    private isModerator = false;
    private isFrozen = false;
    private isLocked = false;
    private stickyCommentHex = 'none';
    private authMethods: StringBooleanMap = {};
    private anonymousOnly = false;
    private sortPolicy: SortPolicy = 'score-desc';
    private selfHex?: string;
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

        // Begin by loading the stylesheet
        await this.cssLoad(`${this.cdn}/css/comentario.css`);

        // Load stylesheet override, if any
        if (this.cssOverride) {
            await this.cssLoad(this.cssOverride);
        }

        // Set up the root content
        this.root
            .classes('root', !this.noFonts && 'root-font')
            .append(
                // Profile bar
                this.profileBar = new ProfileBar(
                    this.cdn,
                    this.origin,
                    this.root,
                    (email, password) => this.authenticateLocally(email, password),
                    idp => this.openOAuthPopup(idp),
                    (name, website, email, password) => this.signup(name, website, email, password)),
                // Main area
                this.mainArea = UIToolkit.div('main-area'),
                // Footer
                UIToolkit.div('footer')
                    .append(
                        UIToolkit.div('logo-container')
                            .append(
                                Wrap.new('a')
                                    .attr({href: 'https://comentario.app/', target: '_blank'})
                                    .html('Powered by ')
                                    .append(Wrap.new('span').classes('logo-brand').inner('Comentario')))));

        // Load information about ourselves
        await this.getAuthStatus();

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
     * Reload the app UI.
     */
    private async reload() {
        // Fetch page data and comments
        await this.loadPageData();

        // Update the main area
        this.setupMainArea();

        // Render the comments
        this.renderComments();
    }

    /**
     * Read page settings from the data-* tags on the comentario script node.
     */
    private dataTagsLoad() {
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
            Wrap.byId(`card-${id}`)
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
            .append(...new CommentTree().render(this.makeCommentRenderingContext(), 'root'));
    }

    /**
     * Set and display (message is given) or clean (message is falsy) an error message in the error panel.
     * @param message Message to set. If falsy, the error panel gets removed.
     * @private
     */
    private setError(message?: string) {
        if (message) {
            if (!this.error) {
                this.root.prepend(this.error = UIToolkit.div('error-box'));
            }
            this.error.inner(message);
        } else {
            this.error?.remove();
            this.error = undefined;
        }
    }

    /**
     * Check the provided response for error status and set the error message if it is, otherwise remove the error panel.
     * @param response Response to check.
     * @return Whether there was an error.
     * @private
     */
    private checkError(response: ApiResponseBase): boolean {
        this.setError(!response.success && response.message);
        return !response.success;
    }

    /**
     * Request the authentication status of the current user from the backend, and return a promise that resolves as
     * soon as the status becomes definite.
     * @private
     */
    private async getAuthStatus(): Promise<void> {
        this.isAuthenticated = false;
        this.isModerator = false;
        this.selfHex = undefined;

        // If we're not already (knowingly) anonymous
        const token = this.token;
        if (token !== 'anonymous') {
            // Fetch the status from the backend
            try {
                const r = await this.apiClient.post<ApiSelfResponse>('commenter/self', {commenterToken: token});
                if (!r.success) {
                    this.token = 'anonymous';
                } else {
                    this.profileBar.authenticated(r.commenter, r.email, token, () => this.logout());
                    this.isAuthenticated = true;

                    // Store ourselves' data as commenter data
                    this.commenters[r.commenter.commenterHex] = r.commenter;
                    this.selfHex = r.commenter.commenterHex;
                }
            } catch (e) {
                // On any error consider the user unauthenticated
                console.error(e);
            }
        }

        // Clean up the profile bar in the case the user isn't authenticated (known auth methods will be set up later)
        if (!this.isAuthenticated) {
            this.profileBar.notAuthenticated();
        }
    }

    /**
     * Create and return a main area element.
     * @private
     */
    private setupMainArea() {
        // Clean up everything from the main area
        this.mainArea.html('');
        this.modTools = null;
        this.modToolsLockBtn = null;
        this.commentsArea = null;

        // Add a moderator toolbar, in necessary
        if (this.isModerator) {
            this.mainArea.append(
                this.modTools = UIToolkit.div('mod-tools')
                    .append(
                        // Title
                        Wrap.new('span').classes('mod-tools-title').inner('Moderator tools'),
                        // Lock/Unlock button
                        this.modToolsLockBtn = UIToolkit.button(
                            this.isLocked ? 'Unlock thread' : 'Lock thread',
                            () => this.threadLockToggle())));
        }

        // If commenting is locked/frozen, add a corresponding message
        if (this.isLocked || this.isFrozen) {
            this.mainArea.append(UIToolkit.div('moderation-notice').inner('This thread is locked. You cannot add new comments.'));

        // Otherwise, add a comment editor host, which will get an editor for creating a new comment
        } else {
            this.mainArea.append(
                this.addCommentHost = UIToolkit.div('add-comment-host')
                    .attr({tabindex: '0'})
                    .on('focus', () => this.addComment(null)));
        }

        // If there's any comment, add sort buttons
        if (this.parentHexMap) {
            this.mainArea.append(new SortBar(
                sp => {
                    this.sortPolicy = sp;
                    // Re-render comments using the new sort
                    this.renderComments();
                },
                this.sortPolicy));
        }

        // Create a panel for comments
        this.commentsArea = UIToolkit.div('comments').appendTo(this.mainArea);
    }

    /**
     * Start editing new comment.
     * @param parentCard Parent card for adding a reply to. If falsy, a top-level comment is being added
     * @private
     */
    private addComment(parentCard: CommentCard) {
        // Kill any existing editor
        this.cancelCommentEdits();

        // Create a new editor
        this.editor = new CommentEditor(
            parentCard?.children || this.addCommentHost,
            this.root,
            false,
            '',
            this.requireIdentification,
            this.anonymousOnly,
            () => this.cancelCommentEdits(),
            editor => this.submitNewComment(parentCard, editor.markdown, editor.anonymous));
    }

    /**
     * Start editing existing comment.
     * @param card Card hosting the comment.
     * @private
     */
    private editComment(card: CommentCard) {
        // Kill any existing editor
        this.cancelCommentEdits();

        // Create a new editor
        this.editor = new CommentEditor(
            card,
            this.root,
            true,
            card.comment.markdown,
            true,
            false,
            () => this.cancelCommentEdits(),
            editor => this.submitCommentEdits(card, editor.markdown));
    }

    /**
     * Submit a new comment to the backend, authenticating the user before if necessary.
     * @param parentCard Parent card for adding a reply to. If falsy, a top-level comment is being added
     * @param markdown Markdown text entered by the user.
     * @param anonymous Whether the user chose to comment anonymously.
     * @private
     */
    private async submitNewComment(parentCard: CommentCard, markdown: string, anonymous: boolean): Promise<void> {
        // Authenticate the user, if required
        const auth = this.requireIdentification || !anonymous;
        if (!this.isAuthenticated && auth) {
            await this.profileBar.loginUser();
        }

        // If we can proceed: user logged in or that wasn't required
        if (this.isAuthenticated || !auth) {
            // Submit the comment to the backend
            const commenterToken = this.isAuthenticated ? this.token : 'anonymous';
            const r = await this.apiClient.post<ApiCommentNewResponse>('comment/new', {
                commenterToken,
                domain:    parent.location.host,
                path:      this.pageId,
                parentHex: parentCard?.comment.commentHex || 'root',
                markdown,
            });
            if (this.checkError(r)) {
                return;
            }

            // Add a new comment card
            const comment: Comment = {
                commentHex:   r.commentHex,
                commenterHex: this.selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : this.selfHex,
                markdown,
                html:         r.html,
                parentHex:    'root',
                score:        0,
                state:        r.state,
                direction:    0,
                creationDate: new Date().toISOString(),
                deleted:      false,
            };
            const newCard = new CommentCard(comment, this.makeCommentRenderingContext({}));

            // Adding a reply
            if (parentCard) {
                parentCard.prependCard(newCard);

            } else {
                // Adding a top-level comment
                this.commentsArea.prepend(newCard);
            }

            // Remove the editor
            this.cancelCommentEdits();
        }
    }

    /**
     * Submit the entered comment markdown to the backend for saving.
     * @param card Card whose comment is being updated.
     * @param markdown Markdown text entered by the user.
     */
    private async submitCommentEdits(card: CommentCard, markdown: string): Promise<void> {
        // Submit the edit to the backend
        const r = await this.apiClient.post<ApiCommentEditResponse>(
            'comment/edit',
            {commenterToken: this.token, commentHex: card.comment.commentHex, markdown});
        if (this.checkError(r)) {
            return;
        }

        // Update the locally stored comment's data
        card.comment.markdown = markdown;
        card.comment.html = r.html;

        // Update the state of the card and its text
        card.update();
        card.updateText();

        // Remove the editor
        this.cancelCommentEdits();
    }

    /**
     * Stop editing comment and remove any existing editor.
     * @private
     */
    private cancelCommentEdits() {
        this.editor?.remove();
    }

    /**
     * Register the user with the given details and log them in.
     * @param name User's full name.
     * @param website User's website.
     * @param email User's email.
     * @param password User's password.
     */
    private async signup(name: string, website: string, email: string, password: string): Promise<void> {
        // Sign the user up
        const r = await this.apiClient.post<ApiResponseBase>('commenter/new', {name, website, email, password});
        if (this.checkError(r)) {
            return Promise.reject();
        }

        // Log the user in
        return this.authenticateLocally(email, password);
    }

    /**
     * Authenticate the user using local authentication (email and password).
     * @param email User's email.
     * @param password User's password.
     */
    private async authenticateLocally(email: string, password: string): Promise<void> {
        // Log the user in
        const r = await this.apiClient.post<ApiCommenterLoginResponse>('commenter/login', {email, password});
        if (this.checkError(r)) {
            return Promise.reject();
        }

        // Store the authenticated token in a cookie
        this.token = r.commenterToken;

        // Refresh the auth status
        return this.getAuthStatus();
    }

    /**
     * Open a new browser popup window for authenticating with the given identity provider and return a promise that
     * resolves as soon as the user is authenticated, or rejects when the authentication has been unsuccessful.
     * @param idp Identity provider to initiate authentication with.
     * @private
     */
    private async openOAuthPopup(idp: string): Promise<void> {
        // Request a token
        const r = await this.apiClient.get<ApiCommenterTokenNewResponse>('commenter/token/new');
        if (this.checkError(r)) {
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
        return this.getAuthStatus();
    }

    /**
     * Log the current user out.
     * @private
     */
    private async logout(): Promise<void> {
        this.token = 'anonymous';
        await this.getAuthStatus();
        return this.reload();
    }

    /**
     * Load data for the current page URL, including the comments, from the backend and store them locally
     * @private
     */
    private async loadPageData(): Promise<void> {
        // Retrieve page settings and a comment list from the backend
        const r = await this.apiClient.post<ApiCommentListResponse>('comment/list', {
            commenterToken: this.token,
            domain:         parent.location.host,
            path:           this.pageId,
        });
        if (this.checkError(r)) {
            return;
        }

        // Store page- and backend-related properties
        this.requireIdentification = r.requireIdentification;
        this.isModerator           = r.isModerator;
        this.isFrozen              = r.isFrozen;
        this.isLocked              = r.attributes.isLocked;
        this.stickyCommentHex      = r.attributes.stickyCommentHex;
        this.authMethods           = r.configuredOauths;
        this.sortPolicy            = r.defaultSortPolicy;

        // Check if no auth provider available, but we allow anonymous commenting
        this.anonymousOnly = !this.requireIdentification && !Object.values(this.authMethods).includes(true);

        // Configure methods in the profile bar
        this.profileBar.authMethods = this.authMethods;

        // Build a map by grouping all comments by their parentHex value
        this.parentHexMap = r.comments.reduce(
            (m, c) => {
                // Also calculate each comment's creation time in milliseconds
                c.creationMs = new Date(c.creationDate).getTime();
                const ph = c.parentHex;
                if (ph in m) {
                    m[ph].push(c);
                } else {
                    m[ph] = [c];
                }
                return m;
            },
            {} as CommentsGroupedByHex);

        // Store all known commenters
        Object.assign(this.commenters, r.commenters);
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
    private async approveComment(card: CommentCard): Promise<void> {
        // Submit the approval to the backend
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/approve',
            {commenterToken: this.token, commentHex: card.comment.commentHex});
        if (this.checkError(r)) {
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
    private async deleteComment(card: CommentCard): Promise<void> {
        // Run deletion with the backend
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/delete',
            {commenterToken: this.token, commentHex: card.comment.commentHex});
        if (this.checkError(r)) {
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
    private async stickyComment(card: CommentCard): Promise<void> {
        // Save the page's sticky comment ID
        this.stickyCommentHex = this.stickyCommentHex === card.comment.commentHex ? 'none' : card.comment.commentHex;
        await this.submitPageAttrs();

        // Reload all comments
        return this.reload();
    }

    /**
     * Vote (upvote, downvote, or undo vote) for the given comment.
     * @private
     */
    private async voteComment(card: CommentCard, direction: -1 | 0 | 1): Promise<void> {
        // Only registered users can vote
        if (!this.isAuthenticated) {
            await this.profileBar.loginUser();

            // Failed to authenticate
            if (!this.isAuthenticated) {
                return;
            }
        }

        // Run the vote with the API
        const r = await this.apiClient.post<ApiResponseBase>(
            'comment/vote',
            {commenterToken: this.token, commentHex: card.comment.commentHex, direction});
        if (this.checkError(r)) {
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
        this.checkError(r);
    }

    /**
     * Return a new comment rendering context.
     * @param parentMap Optional parent map to use. If not provided, uses the map for all comments on the page (parentHexMap).
     */
    private makeCommentRenderingContext(parentMap?: CommentsGroupedByHex): CommentRenderingContext {
        return {
            cdn:             this.cdn,
            root:            this.root,
            parentMap:       parentMap || this.parentHexMap,
            commenters:      this.commenters,
            selfHex:         this.selfHex,
            stickyHex:       this.stickyCommentHex,
            sortPolicy:      this.sortPolicy,
            isAuthenticated: this.isAuthenticated,
            isModerator:     this.isModerator,
            isLocked:        this.isLocked || this.isFrozen,
            hideDeleted:     this.hideDeleted,
            curTimeMs:       new Date().getTime(),
            onApprove:       card => this.approveComment(card),
            onDelete:        card => this.deleteComment(card),
            onEdit:          card => this.editComment(card),
            onReply:         card => this.addComment(card),
            onSticky:        card => this.stickyComment(card),
            onVote:          (card, direction) => this.voteComment(card, direction),
        };
    }
}
