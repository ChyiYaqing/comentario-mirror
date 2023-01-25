import { HttpClient } from './http-client';
import {
    Comment,
    Commenter,
    CommentMap,
    CommentsGroupedByHex,
    Email,
    SortPolicy,
    SortPolicyProps,
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

const IDS = {
    mainArea:           'main-area',
    login:              'login',
    modTools:           'mod-tools',
    modToolsLockButton: 'mod-tools-lock-button',
    error:              'error',
    loggedContainer:    'logged-container',
    preCommentsArea:    'pre-comments-area',
    commentsArea:       'comments-area',
    superContainer:     'textarea-super-container-',
    textarea:           'textarea-',
    anonymousCheckbox:  'anonymous-checkbox-',
    sortPolicy:         'sort-policy-',
    card:               'comment-card-',
    body:               'comment-body-',
    text:               'comment-text-',
    score:              'comment-score-',
    edit:               'comment-edit-',
    reply:              'comment-reply-',
    collapse:           'comment-collapse-',
    upvote:             'comment-upvote-',
    downvote:           'comment-downvote-',
    approve:            'comment-approve-',
    sticky:             'comment-sticky-',
    children:           'comment-children-',
    name:               'comment-name-',
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
    private rootId = 'commento';

    private root: Wrap<any>;
    private pageId = parent.location.pathname;
    private cssOverride: string;
    private noFonts = false;
    private hideDeleted = false;
    private autoInit = true;
    private isAuthenticated = false;
    private comments: Comment[] = [];

    /** Loaded comment objects indexed by commentHex. */
    private commentsByHex: CommentMap = {};

    private readonly commenters: { [k: string]: Commenter } = {};
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
        this.errorElementCreate();

        // Load information about ourselves
        return this.selfGet()
            // Fetch comments
            .then(() => this.loadComments())
            // Create the layout
            .then(() => {
                this.modToolsCreate();
                this.mainAreaCreate();
                this.rootCreate();
                this.commentsRender();
                this.root.append(this.footerLoad());
                this.scrollToCommentHash();
                this.allShow();
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
        return (this.autoInit ? this.main() : Promise.resolve())
            .then(() => console.info(`Initialised Comentario ${this.version}`));
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
                        Wrap.new('link').attr({href: url, rel: 'stylesheet', type: 'text/css'}).on('load', () => resolve()));
            });
    }

    footerLoad(): Wrap<HTMLDivElement> {
        return Wrap.new('div')
            .classes('footer')
            .append(
                Wrap.new('div')
                    .classes('logo-container')
                    .append(
                        Wrap.new('a')
                            .attr({href: 'https://comentario.app/', target: '_blank'})
                            .html('Powered by ')
                            .append(
                                Wrap.new('span').classes('logo'),
                                Wrap.new('span').classes('logo-brand').inner('Comentario'))));
    }

    loadComments(): Promise<void> {
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

                // Update comment models and make a hex-comment map
                this.commentsByHex = {};
                this.comments.forEach(c => {
                    c.creationMs = new Date(c.creationDate).getTime();
                    this.commentsByHex[c.commentHex] = c;
                });
            });
    }

    errorShow(text: string) {
        Wrap.byId(IDS.error).inner(text).style('display: block;');
    }

    errorHide() {
        Wrap.byId(IDS.error).style('display: none;');
    }

    errorElementCreate() {
        Wrap.new('div').id(IDS.error).classes('error-box').style('display: none;').appendTo(this.root);
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
            anonContainer = Wrap.new('div')
                .classes('round-check', 'anonymous-checkbox-container')
                .append(
                    anonCheckbox,
                    Wrap.new('label').attr({for: Wrap.idPrefix + IDS.anonymousCheckbox + commentHex}).inner('Comment anonymously'));
        }

        // Instantiate and set up a new form
        return UIToolkit.form(() => isEdit ? this.saveCommentEdits(commentHex) : this.submitAccountDecide(commentHex))
            .id(IDS.superContainer + commentHex)
            .classes('textarea-form')
            .append(
                // Textarea in a container
                Wrap.new('div')
                    .classes('textarea-container')
                    .append(
                        Wrap.new('textarea').id(IDS.textarea + commentHex).attr({placeholder: 'Add a comment'}).autoExpand()),
                // Textarea footer
                Wrap.new('div')
                    .classes('textarea-form-footer')
                    .append(
                        Wrap.new('div')
                            .append(
                                // Anonymous checkbox, if any
                                anonContainer,
                                // Markdown help button
                                Wrap.new('a')
                                    .classes('markdown-button')
                                    .html('<b>M⬇</b>&nbsp;Markdown')
                                    .click(() => MarkdownHelp.run(this.root))),
                        // Save button
                        Wrap.new('button')
                            .attr({type: 'submit'})
                            .classes('button', 'submit-button')
                            .inner(isEdit ? 'Save Changes' : 'Add Comment')));
    }

    sortPolicyApply(policy: SortPolicy) {
        Wrap.byId(IDS.sortPolicy + this.sortPolicy).noClasses('sort-policy-button-selected');
        Wrap.byId(IDS.sortPolicy + policy).classes('sort-policy-button-selected');
        this.sortPolicy = policy;

        // Re-render the sorted comment
        this.commentsRender();
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
        const loginText       = Wrap.new('div').classes('login-text').inner('Login').click(() => this.showLoginDialog(null));
        const preCommentsArea = Wrap.new('div').id(IDS.preCommentsArea);
        const commentsArea    = Wrap.new('div').id(IDS.commentsArea).classes('comments');

        // If there's any auth provider configured, add a Login button
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
        mainArea.append(preCommentsArea, commentsArea);
    }

    messageCreate(text: string): Wrap<HTMLDivElement> {
        return Wrap.new('div').classes('moderation-notice').inner(text);
    }

    commentNew(commentHex: string, commenterToken: string, appendCard: boolean): Promise<void> {
        const container   = Wrap.byId(IDS.superContainer + commentHex);
        const textarea    = Wrap.byId(IDS.textarea + commentHex);

        // Validate the textarea value
        const markdown = textarea.val;
        if (markdown === '') {
            textarea.classes('red-border');
            return Promise.reject();
        }

        textarea.noClasses('red-border');

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
                    this.messageCreate(message).prependTo(Wrap.byId(IDS.superContainer + commentHex));
                }

                const comment: Comment = {
                    commentHex:   resp.commentHex,
                    commenterHex: this.selfHex === undefined || commenterToken === 'anonymous' ? 'anonymous' : this.selfHex,
                    markdown,
                    html:         resp.html,
                    parentHex:    'root',
                    score:        0,
                    state:        'approved',
                    direction:    0,
                    creationDate: new Date().toISOString(),
                    deleted:      false,
                };

                const newCard = this.commentsRecurse({root: [comment]}, 'root');

                // Store the updated comment in the local map
                this.commentsByHex[resp.commentHex] = comment;
                if (appendCard) {
                    if (commentHex !== 'root') {
                        container.replaceWith(newCard);

                        this.shownReply[commentHex] = false;
                        Wrap.byId(IDS.reply + commentHex)
                            .noClasses('option-cancel')
                            .classes('option-reply')
                            .attr({title: 'Reply to this comment'})
                            .click(() => this.replyShow(commentHex));
                    } else {
                        textarea.value('');
                        newCard.insertAfter(Wrap.byId(IDS.preCommentsArea));
                    }
                } else if (commentHex === 'root') {
                    textarea.value('');
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

    commentsRecurse(parentMap: CommentsGroupedByHex, parentHex: string): Wrap<any> {
        // Fetch comments that have the given parentHex
        const comments = parentMap[parentHex];

        // Return an empty wrap if there's none
        if (!comments?.length) {
            return new Wrap();
        }

        // Apply the chosen sorting, always keeping the sticky comment on top
        comments.sort((a, b) =>
            !a.deleted && a.commentHex === this.stickyCommentHex ?
                -Infinity :
                !b.deleted && b.commentHex === this.stickyCommentHex ?
                    Infinity :
                    this.sortingProps[this.sortPolicy].comparator(a, b));

        const curTime = (new Date()).getTime();
        const cards = Wrap.new('div');
        comments.forEach(comment => {
            const commenter = this.commenters[comment.commenterHex];
            const commHasLink = commenter.link && commenter.link !== 'undefined' && commenter.link !== 'https://undefined';
            const hex = comment.commentHex;
            const color = this.colorGet(`${comment.commenterHex}-${commenter.name}`);
            const children = this.commentsRecurse(parentMap, hex).id(IDS.children + hex).classes('body');
            const card = Wrap.new('div')
                .id(IDS.card + hex)
                .style(`border-left: 2px solid ${color}`)
                .classes('card', this.isModerator && comment.state !== 'approved' && 'dark-card')
                .append(
                    // Card header
                    Wrap.new('div')
                        .classes('header')
                        .append(
                            // Options toolbar
                            this.getCommentOptions(comment, hex, parentHex),
                            // Avatar
                            commenter.photo === 'undefined' ?
                                Wrap.new('div')
                                    .style(`background-color: ${color}`)
                                    .classes('avatar')
                                    .html(comment.commenterHex === 'anonymous' ? '?' : commenter.name[0].toUpperCase()) :
                                Wrap.new('img')
                                    .classes('avatar-img')
                                    .attr({src: `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`, alt: ''}),
                            // Name
                            Wrap.new(commHasLink ? 'a' : 'div')
                                .id(IDS.name + hex)
                                .inner(comment.deleted ? '[deleted]' : commenter.name)
                                .classes(
                                    'name',
                                    commenter.isModerator && 'moderator',
                                    comment.state === 'flagged' && 'flagged')
                                .attr(commHasLink && {href: commenter.link, rel: 'nofollow noopener noreferrer'}),
                            // Subtitle
                            Wrap.new('div')
                                .classes('subtitle')
                                .append(
                                    // Score
                                    Wrap.new('div').id(IDS.score + hex).classes('score').inner(this.scorify(comment.score)),
                                    // Time ago
                                    Wrap.new('div')
                                        .classes('timeago')
                                        .inner(this.timeDifference(curTime, comment.creationMs))
                                        .attr({title: comment.creationDate.toString()}))),
                    // Card contents
                    Wrap.new('div')
                        .append(
                            Wrap.new('div').id(IDS.body + hex).classes('body')
                                .append(Wrap.new('div').id(IDS.text + hex).html(comment.html)),
                            children));

            if (!comment.deleted || !this.hideDeleted && !children.ok) {
                cards.append(card);
            }
        });

        // If no cards found, return an empty wrap
        return cards.hasChildren ? cards : new Wrap();
    }

    /**
     * Return a wrapped options toolbar for a comment.
     * @private
     */
    private getCommentOptions(comment: Comment, hex: string, parentHex: string): Wrap<any> {
        const options = Wrap.new('div').classes('options');

        // Sticky comment indicator (for non-moderator only)
        const isSticky = this.stickyCommentHex === hex;
        if (!comment.deleted && !this.isModerator && isSticky) {
            Wrap.new('button')
                .id(IDS.sticky + hex)
                .classes('option-button', 'option-sticky')
                .attr({title: 'This comment has been stickied', type: 'button', disabled: 'true'})
                .appendTo(options);
        }

        // Approve button
        if (this.isModerator && comment.state !== 'approved') {
            Wrap.new('button')
                .id(IDS.approve + hex)
                .classes('option-button', 'option-approve')
                .attr({type: 'button', title: 'Approve'})
                .click(() => this.commentApprove(hex))
                .appendTo(options);
        }

        // Remove button
        if (!comment.deleted && (this.isModerator || comment.commenterHex === this.selfHex)) {
            Wrap.new('button')
                .classes('option-button', 'option-remove')
                .attr({type: 'button', title: 'Remove'})
                .click(() => this.commentDelete(hex))
                .appendTo(options);
        }

        // Sticky toggle button (for moderator and a top-level comments only)
        if (!comment.deleted && this.isModerator && parentHex === 'root') {
            Wrap.new('button')
                .id(IDS.sticky + hex)
                .classes('option-button', isSticky ? 'option-unsticky' : 'option-sticky')
                .attr({title: isSticky ? 'Unsticky' : 'Sticky', type: 'button'})
                .click(() => this.commentSticky(hex))
                .appendTo(options);
        }

        // Own comment: Edit button
        if (comment.commenterHex === this.selfHex) {
            Wrap.new('button')
                .id(IDS.edit + hex)
                .classes('option-button', 'option-edit')
                .attr({type: 'button', title: 'Edit'})
                .click(() => this.startEditing(hex))
                .appendTo(options);

        // Someone other's comment: Reply button
        } else if (!comment.deleted) {
            Wrap.new('button')
                .id(IDS.reply + hex)
                .classes('option-button', 'option-reply')
                .attr({type: 'button', title: 'Reply'})
                .click(() => this.replyShow(hex))
                .appendTo(options);
        }

        // Upvote / Downvote buttons
        if (!comment.deleted) {
            this.updateUpDownAction(
                Wrap.new('button')
                    .id(IDS.upvote + hex)
                    .classes('option-button', 'option-upvote', this.isAuthenticated && comment.direction > 0 && 'upvoted')
                    .attr({type: 'button', title: 'Upvote'})
                    .appendTo(options),
                Wrap.new('button')
                    .id(IDS.downvote + hex)
                    .classes('option-button', 'option-downvote', this.isAuthenticated && comment.direction < 0 && 'downvoted')
                    .attr({type: 'button', title: 'Downvote'})
                    .appendTo(options),
                hex,
                comment.direction);
        }

        // Collapse button
        Wrap.new('button')
            .id(IDS.collapse + hex)
            .classes('option-button', 'option-collapse')
            .attr({type: 'button', title: 'Collapse children'})
            .click(() => this.commentCollapse(hex))
            .appendTo(options);
        return options;
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
                Wrap.byId(IDS.card + commentHex).noClasses('dark-card');
                Wrap.byId(IDS.name + commentHex).noClasses('flagged');
                Wrap.byId(IDS.approve + commentHex).remove();
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
                Wrap.byId(IDS.text + commentHex).inner('[deleted]');
            });
    }

    updateUpDownAction(upvote: Wrap<any>, downvote: Wrap<any>, commentHex: string, direction: number) {
        let oldDir = 0, du = 1, dd = -1;
        if (direction > 0) {
            oldDir = 1;
            du = 0;
            dd = -1;
        } else if (direction < 0) {
            oldDir = -1;
            du = 1;
            dd = 0;
        }
        upvote  .unlisten().click(() => this.isAuthenticated ? this.vote(commentHex, oldDir, du) : this.showLoginDialog(null));
        downvote.unlisten().click(() => this.isAuthenticated ? this.vote(commentHex, oldDir, dd) : this.showLoginDialog(null));
    }

    vote(commentHex: string, oldDirection: number, direction: number): Promise<void> {
        const upvote   = Wrap.byId(IDS.upvote   + commentHex).noClasses('upvoted')  .classes(direction > 0 && 'upvoted');
        const downvote = Wrap.byId(IDS.downvote + commentHex).noClasses('downvoted').classes(direction < 0 && 'downvoted');
        this.updateUpDownAction(upvote, downvote, commentHex, direction);

        // Find the comment by its hex
        const comment = this.commentsByHex[commentHex];
        if (!comment) {
            return Promise.reject();
        }

        // Update the score reading
        const newScore = comment.score - oldDirection + direction;
        const ws = Wrap.byId(IDS.score + commentHex).inner(this.scorify(newScore));

        // Run the vote with the API
        return this.apiClient.post<ApiResponseBase>('comment/vote', {commenterToken: this.commenterTokenGet(), commentHex, direction})
            .then(resp => {
                // Undo the vote on failure
                if (!resp.success) {
                    this.errorShow(resp.message);
                    upvote.noClasses('upvoted');
                    downvote.noClasses('downvoted');
                    ws.inner(this.scorify(comment.score));
                    this.updateUpDownAction(upvote, downvote, commentHex, oldDirection);
                    return Promise.reject();
                }

                // Succeeded
                this.errorHide();
                comment.score = newScore;
                return undefined;
            });
    }

    /**
     * Submit the entered comment markdown to the backend for saving.
     * @param commentHex Comment's hex ID
     */
    saveCommentEdits(commentHex: string): Promise<void> {
        const textarea = Wrap.byId(IDS.textarea + commentHex);
        const markdown = textarea.val.trim();
        if (markdown === '') {
            textarea.classes('red-border');
            return Promise.reject();
        }

        textarea.noClasses('red-border');

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
                    this.messageCreate(message).prependTo(Wrap.byId(IDS.superContainer + commentHex));
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

    commentCollapse(commentHex: string) {
        Wrap.byId(IDS.children + commentHex).classes('hidden');
        Wrap.byId(IDS.collapse + commentHex)
            .noClasses('option-collapse')
            .classes('option-uncollapse')
            .attr({title: 'Expand children'})
            .unlisten()
            .click(() => this.commentUncollapse(commentHex));
    }

    commentUncollapse(commentHex: string) {
        Wrap.byId(IDS.children + commentHex).noClasses('hidden');
        Wrap.byId(IDS.collapse + commentHex)
            .noClasses('option-uncollapse')
            .classes('option-collapse')
            .attr({title: 'Collapse children'})
            .unlisten()
            .click(() => this.commentCollapse(commentHex));
    }

    commentsRender() {
        // Group comments by parent hex ID: make map {parentHex: Comment[]}
        const parentMap = this.comments.reduce(
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

        // Re-render the comment recursively and add them to the comments area
        Wrap.byId(IDS.commentsArea)
            .html('')
            .append(this.commentsRecurse(parentMap, 'root'));
    }

    submitAuthenticated(commentHex: string): Promise<void> {
        if (this.isAuthenticated) {
            return this.commentNew(commentHex, this.commenterTokenGet(), true);
        }

        return this.showLoginDialog(commentHex);
    }

    submitAnonymous(commentHex: string): Promise<void> {
        this.chosenAnonymous = true;
        return this.commentNew(commentHex, 'anonymous', true);
    }

    submitAccountDecide(commentHex: string): Promise<void> {
        if (this.requireIdentification) {
            return this.submitAuthenticated(commentHex);
        }

        const anonCheckbox = Wrap.byId(IDS.anonymousCheckbox + commentHex);
        const textarea = Wrap.byId(IDS.textarea + commentHex);
        if (!textarea.val?.trim()) {
            textarea.classes('red-border');
            return Promise.reject();
        }

        textarea.noClasses('red-border');
        return anonCheckbox.isChecked ? this.submitAnonymous(commentHex) : this.submitAuthenticated(commentHex);
    }

    openOAuthPopup(idp: string, commentHex: string): Promise<void> {
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
                popup.location = `${this.origin}/api/oauth/${idp}/redirect?commenterToken=${resp.commenterToken}`;

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
                if (this.isAuthenticated) {
                    Wrap.byId(IDS.loggedContainer).style(null);

                    // Hide the login button
                    Wrap.byId(IDS.login).remove();

                    // Submit the pending comment, if there was one
                    return commentHex && this.commentNew(commentHex, this.commenterTokenGet(), false);
                }
                return undefined;
            })
            .then(() => this.isAuthenticated && this.loadComments())
            .then(() => this.isAuthenticated && this.commentsRender());
    }

    forgotPassword() {
        window.open(`${this.origin}/forgot?commenter=true`, '_blank');
    }

    authenticateLocally(email: string, password: string, commentHex: string): Promise<void> {
        return this.apiClient.post<ApiCommenterLoginResponse>('commenter/login', {email, password})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return Promise.reject();
                }

                this.errorHide();
                this.cookieSet('commentoCommenterToken', resp.commenterToken);
                this.selfLoad(resp.commenter, resp.email);
                Wrap.byId(IDS.login).remove();
                return commentHex ? this.commentNew(commentHex, resp.commenterToken, false) : undefined;
            })
            .then(() => this.loadComments())
            .then(() => this.commentsRender())
            .then(() => this.allShow());
    }

    signup(name: string, website: string, email: string, password: string, commentHex: string): Promise<void> {
        return this.apiClient.post<ApiResponseBase>('commenter/new', {name, website, email, password})
            .then(resp => {
                if (!resp.success) {
                    this.errorShow(resp.message);
                    return Promise.reject();
                }

                this.errorHide();
                return undefined;
            })
            .then(() => this.authenticateLocally(email, password, commentHex));
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
        this.isLocked = !this.isLocked;
        const lock = Wrap.byId(IDS.modToolsLockButton).attr({disabled: 'true'});
        return this.pageUpdate()
            .then(() => lock.attr({disabled: 'false'}))
            .then(() => this.reload());
    }

    commentSticky(commentHex: string): Promise<void> {
        if (this.stickyCommentHex !== 'none') {
            Wrap.byId(IDS.sticky + this.stickyCommentHex).noClasses('option-unsticky').classes('option-sticky');
        }

        this.stickyCommentHex = this.stickyCommentHex === commentHex ? 'none' : commentHex;

        return this.pageUpdate()
            .then(() =>
                void Wrap.byId(IDS.sticky + commentHex)
                    .noClasses(this.stickyCommentHex === commentHex ? 'option-sticky' : 'option-unsticky')
                    .classes(this.stickyCommentHex === commentHex ? 'option-unsticky' : 'option-sticky'));
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
        Wrap.byId(IDS.mainArea).style(null);
        Wrap.byId(IDS.loggedContainer).style(null);
        if (this.isModerator) {
            Wrap.byId(IDS.modTools).style(null);
        }
    }

    async showLoginDialog(commentHex: string) {
        const dlg = await LoginDialog.run(this.root, this.configuredOauths);
        if (dlg.confirmed) {
            switch (dlg.navigateTo) {
                case null:
                    // Local auth
                    await this.authenticateLocally(dlg.email, dlg.password, commentHex);
                    break;

                case 'forgot':
                    // Navigate to forgot password
                    this.forgotPassword();
                    break;

                case 'signup':
                    // Switch to signup
                    await this.showSignupDialog(commentHex);
                    break;

                default:
                    // External auth
                    await this.openOAuthPopup(dlg.navigateTo, commentHex);
            }
        }
    }

    async showSignupDialog(commentHex: string) {
        const dlg = await SignupDialog.run(this.root);
        if (dlg.confirmed) {
            await this.signup(dlg.name, dlg.website, dlg.email, dlg.password, commentHex);
        }
    }

    dataTagsLoad() {
        for (const script of this.doc.getElementsByTagName('script')) {
            if (script.src.match(/\/js\/commento\.js$/)) {
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

    scrollToCommentHash() {
        const h = window.location.hash;
        if (h?.startsWith('#commento-')) {
            const id = window.location.hash.split('-')[1];
            const el = Wrap.byId(IDS.card + id);
            if (!el.ok) {
                // A hack to make sure it's a valid ID before showing the user a message.
                if (id.length === 64) {
                    this.errorShow('The comment you\'re looking for doesn\'t exist; possibly it was deleted.');
                }
                return;
            }

            el.classes('highlighted-card').scrollTo();
        } else if (h?.startsWith('#commento')) {
            this.root.scrollTo();
        }
    }
}
