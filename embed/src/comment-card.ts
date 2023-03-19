import { Wrap } from './element-wrap';
import { AnonymousCommenterId, Comment, CommenterMap, CommentsGroupedByHex, sortingProps, SortPolicy } from './models';
import { UIToolkit } from './ui-toolkit';
import { Utils } from './utils';
import { ConfirmDialog } from './confirm-dialog';

export type CommentCardEventHandler = (c: CommentCard) => void;
export type CommentCardVoteEventHandler = (c: CommentCard, direction: -1 | 0 | 1) => void;

/**
 * Context for rendering comment trees.
 */
export interface CommentRenderingContext {
    /** Base API URL. */
    readonly apiUrl: string;
    /** The root element (for displaying popups). */
    readonly root: Wrap<any>;
    /** Map that links comment lists to their parent hex ID. */
    readonly parentMap: CommentsGroupedByHex;
    /** Map of known commenters. */
    readonly commenters: CommenterMap;
    /** Optional hex ID of the current commenter. */
    readonly selfHex?: string;
    /** Optional hex ID of the current sticky comment. */
    readonly stickyHex?: string;
    /** Current sorting. */
    readonly sortPolicy: SortPolicy;
    /** Whether the current user is authenticated. */
    readonly isAuthenticated: boolean;
    /** Whether the current user is a moderator. */
    readonly isModerator: boolean;
    /** Whether comment thread is locked on this page. */
    readonly isLocked: boolean;
    /** Whether to hide deleted comments. */
    readonly hideDeleted: boolean;
    /** Current time in milliseconds. */
    readonly curTimeMs: number;

    // Events
    readonly onApprove: CommentCardEventHandler;
    readonly onDelete: CommentCardEventHandler;
    readonly onEdit: CommentCardEventHandler;
    readonly onReply: CommentCardEventHandler;
    readonly onSticky: CommentCardEventHandler;
    readonly onVote: CommentCardVoteEventHandler;
}

/**
 * A tree structure of comment cards.
 */
export class CommentTree {

    /**
     * Render a branch of comments that all relate to the same given parent.
     */
    render(ctx: CommentRenderingContext, parentHex: string): CommentCard[] {
        // Fetch comments that have the given parentHex
        const comments = ctx.parentMap[parentHex] || [];

        // Apply the chosen sorting, always keeping the sticky comment on top
        comments.sort((a, b) =>
            !a.deleted && a.commentHex === ctx.stickyHex ?
                -Infinity :
                !b.deleted && b.commentHex === ctx.stickyHex ?
                    Infinity :
                    sortingProps[ctx.sortPolicy].comparator(a, b));

        // Render child comments, if any
        return comments
            // Filter out deleted comment, if they're to be hidden
            .filter(c => !ctx.hideDeleted || !c.deleted)
            // Render a comment card
            // eslint-disable-next-line @typescript-eslint/no-use-before-define
            .map(c => new CommentCard(c, ctx));
    }
}

/**
 * Comment card represents an individual comment in the UI.
 */
export class CommentCard extends Wrap<HTMLDivElement> {

    /** Child cards container. Also used to host a reply editor. */
    children?: Wrap<HTMLDivElement>;

    private eName?: Wrap<HTMLDivElement | HTMLAnchorElement>;
    private eScore?: Wrap<HTMLDivElement>;
    private eHeader?: Wrap<HTMLDivElement>;
    private eBody?: Wrap<HTMLDivElement>;
    private eModNotice?: Wrap<HTMLDivElement>;
    private btnApprove?: Wrap<HTMLButtonElement>;
    private btnCollapse?: Wrap<HTMLButtonElement>;
    private btnDelete?: Wrap<HTMLButtonElement>;
    private btnDownvote?: Wrap<HTMLButtonElement>;
    private btnEdit?: Wrap<HTMLButtonElement>;
    private btnReply?: Wrap<HTMLButtonElement>;
    private btnSticky?: Wrap<HTMLButtonElement>;
    private btnUpvote?: Wrap<HTMLButtonElement>;
    private collapsed = false;

    constructor(
        readonly comment: Comment,
        ctx: CommentRenderingContext,
    ) {
        super(UIToolkit.div().element);

        // Render the content
        this.render(ctx);

        // Update the card controls/text
        this.update();
        this.updateText();
    }

    /**
     * Current comment's flagged state.
     */
    get flagged(): boolean {
        return this.comment.state === 'flagged';
    }

    /**
     * Update comment controls according to the related comment's properties.
     */
    update() {
        const c = this.comment;

        // If the comment is deleted
        if (c.deleted) {
            // Remove comment text
            this.eBody?.inner('[deleted]');

            // Remove children
            this.children?.remove();

            // Remove all option buttons
            this.eScore?.remove();
            this.btnApprove?.remove();
            this.btnCollapse?.remove();
            this.btnDelete?.remove();
            this.btnDownvote?.remove();
            this.btnEdit?.remove();
            this.btnReply?.remove();
            this.btnSticky?.remove();
            this.btnUpvote?.remove();
            return;
        }

        // Score
        this.eScore
            ?.inner(c.score?.toString() || '0')
            .setClasses(c.score > 0, 'upvoted').setClasses(c.score < 0, 'downvoted');
        this.btnUpvote?.setClasses(c.direction > 0, 'upvoted');
        this.btnDownvote?.setClasses(c.direction < 0, 'downvoted');

        // Collapsed
        this.btnCollapse
            ?.attr({title: this.collapsed ? 'Expand children' : 'Collapse children'})
            .setClasses(this.collapsed, 'collapsed');

        // Approved
        const flagged = this.flagged;
        this.setClasses(flagged, 'dark-card');
        this.eName?.setClasses(flagged, 'flagged');
        if (!flagged && this.btnApprove) {
            // Remove the Approve button if the comment is approved
            this.btnApprove.remove();
            this.btnApprove = undefined;
        }

        // Moderation notice
        let mn: string | undefined;
        switch (c.state) {
            case 'unapproved':
                mn = 'Your comment is under moderation.';
                break;
            case 'flagged':
                mn = 'Your comment was flagged as spam and is under moderation.';
                break;
        }
        if (mn) {
            // If there's something to display, make sure the notice element exists and appended to the header
            if (!this.eModNotice) {
                this.eModNotice = UIToolkit.div('moderation-notice').appendTo(this.eHeader!);
            }
            this.eModNotice.inner(mn);

        // No moderation notice
        } else if (this.eModNotice) {
            this.eModNotice.remove();
            this.eModNotice = undefined;
        }
    }

    /**
     * Update the current comment's text.
     */
    updateText() {
        this.eBody!.html(this.comment.html || '');
    }

    /**
     * Render the content of the card.
     * @private
     */
    private render(ctx: CommentRenderingContext): void {
        const hex = this.comment.commentHex;
        const commenter = ctx.commenters[this.comment.commenterHex];
        const anonymous = this.comment.commenterHex === AnonymousCommenterId;

        // Figure out if the commenter has a profile link
        const commLink = !commenter.link ? undefined : commenter.link;

        // Pick a color for the commenter
        const bgColor = anonymous ? 'anonymous' : Utils.colourIndex(`${this.comment.commenterHex}-${commenter.name}`);

        // Render children
        this.children = UIToolkit.div('card-children').append(...new CommentTree().render(ctx, hex));

        // Render a card
        this.id(`card-${hex}`) // ID for scrolling to
            .classes('card', `border-${bgColor}`)
            .append(
                // Card header
                this.eHeader = UIToolkit.div('card-header')
                    .append(
                        // Avatar
                        !anonymous && commenter.photo ?
                            Wrap.new('img')
                                .classes('avatar-img')
                                .attr({
                                    src: `${ctx.apiUrl}/commenter/photo?commenterHex=${commenter.commenterHex}`,
                                    alt: '',
                                }) :
                            UIToolkit.div('avatar', `bg-${bgColor}`).html(anonymous ? '' : commenter.name![0].toUpperCase()),
                        // Name and subtitle
                        UIToolkit.div('name-container')
                            .append(
                                // Name
                                this.eName = Wrap.new(commLink ? 'a' : 'div')
                                    .inner(this.comment.deleted ? '[deleted]' : commenter.name!)
                                    .classes('name', commenter.isModerator && 'moderator')
                                    .attr({href: commLink, rel: commLink && 'nofollow noopener noreferrer'}),
                                // Subtitle
                                UIToolkit.div('subtitle')
                                    // Time ago
                                    .inner(Utils.timeAgo(ctx.curTimeMs, this.comment.creationMs!))
                                    .attr({title: this.comment.creationDate.toString()}))),
                // Card body
                this.eBody = UIToolkit.div('card-body'),
                // Options toolbar
                this.commentOptionsBar(ctx, hex, this.comment.parentHex),
                // Children (if any)
                this.children);
    }

    /**
     * Return a wrapped options toolbar for a comment.
     * @private
     */
    private commentOptionsBar(ctx: CommentRenderingContext, hex: string, parentHex: string): Wrap<HTMLDivElement> | null {
        if (this.comment.deleted) {
            return null;
        }
        const options = UIToolkit.div('options');

        // Left- and right-hand side of the options bar
        const left = UIToolkit.div('options-sub').appendTo(options);
        const right = UIToolkit.div('options-sub').appendTo(options);

        // Upvote / Downvote buttons and the score
        left.append(
            this.btnUpvote = this.getOptionButton('upvote', null, () => ctx.onVote(this, this.comment.direction > 0 ? 0 : 1)),
            this.eScore = UIToolkit.div('score').attr({title: 'Comment score'}),
            this.btnDownvote = this.getOptionButton('downvote', null, () => ctx.onVote(this, this.comment.direction < 0 ? 0 : -1)));
        // Reply button
        if (!ctx.isLocked) {
            this.btnReply = this.getOptionButton('reply', null, () => ctx.onReply(this)).appendTo(left);
        }

        // Approve button
        if (ctx.isModerator && this.comment.state !== 'approved') {
            this.btnApprove = this.getOptionButton('approve', 'text-success', () => ctx.onApprove(this)).appendTo(right);
        }

        // Sticky toggle button (top-level comments only). The sticky status can only be changed after a full tree
        // reload
        const isSticky = ctx.stickyHex === hex;
        if (parentHex === 'root' && (isSticky || ctx.isModerator)) {
            this.btnSticky = this.getOptionButton('sticky', null, () => ctx.onSticky(this))
                .setClasses(isSticky, 'text-warning')
                .attr({
                    disabled: ctx.isModerator ? null : 'true',
                    title: isSticky ? (ctx.isModerator ? 'Unsticky' : 'This comment has been stickied') : 'Sticky',
                })
                .appendTo(right);
        }

        // Moderator or own comment
        if (ctx.isModerator || this.comment.commenterHex === ctx.selfHex) {
            right.append(
                // Edit button
                this.btnEdit = this.getOptionButton('edit', null, () => ctx.onEdit(this)).appendTo(right),
                // Delete button
                this.btnDelete = this.getOptionButton('delete', 'text-danger', btn => this.deleteComment(btn, ctx)).appendTo(right));
        }

        // Collapse button, if there are any children
        if (this.children?.hasChildren) {
            this.btnCollapse = this.getOptionButton('collapse', 'btn-collapse', () => this.collapse(!this.collapsed)).appendTo(right);
        }
        return options;
    }

    private async deleteComment(btn: Wrap<any>, ctx: CommentRenderingContext) {
        // Confirm deletion
        if (await ConfirmDialog.run(ctx.root, {ref: btn, placement: 'bottom-end'}, 'Are you sure you want to delete this comment?')) {
            // Notify the callback
            ctx.onDelete(this);
        }
    }

    private collapse(c: boolean) {
        if (this.children?.ok) {
            this.collapsed = c;
            this.children
                .noClasses('fade-in', 'fade-out', !c && 'hidden')
                .on('animationend', ch => ch.classes(c && 'hidden'), true)
                .classes(c && 'fade-out', !c && 'fade-in');
            this.update();
        }
    }

    /**
     * Return a rendered, wrapped option button.
     * @param icon Name of the icon to put on the button.
     * @param cls Optional additional class.
     * @param onClick Button's click handler.
     * @private
     */
    private getOptionButton(icon: 'approve' | 'collapse' | 'delete' | 'downvote' | 'edit' | 'reply' | 'sticky' | 'upvote', cls?: string | null, onClick?: (btn: Wrap<HTMLButtonElement>) => void): Wrap<any> {
        let title: string;
        let svg: string;
        switch (icon) {
            case 'approve':
                title = 'Approve';
                svg = UIToolkit.svg(512, 512, '<path fill="currentColor" d="M470.6 105.4c12.5 12.5 12.5 32.8 0 45.3l-256 256c-12.5 12.5-32.8 12.5-45.3 0l-128-128c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0L192 338.7 425.4 105.4c12.5-12.5 32.8-12.5 45.3 0z"/>');
                break;
            case 'collapse':
                title = 'Collapse';
                svg = UIToolkit.svg(320, 512, '<path fill="currentColor" d="M137.4 374.6c12.5 12.5 32.8 12.5 45.3 0l128-128c9.2-9.2 11.9-22.9 6.9-34.9s-16.6-19.8-29.6-19.8L32 192c-12.9 0-24.6 7.8-29.6 19.8s-2.2 25.7 6.9 34.9l128 128z"/>');
                break;
            case 'delete':
                title = 'Delete';
                svg = UIToolkit.svg(448, 512, '<path fill="currentColor" d="M135.2 17.7L128 32H32C14.3 32 0 46.3 0 64S14.3 96 32 96H416c17.7 0 32-14.3 32-32s-14.3-32-32-32H320l-7.2-14.3C307.4 6.8 296.3 0 284.2 0H163.8c-12.1 0-23.2 6.8-28.6 17.7zM416 128H32L53.2 467c1.6 25.3 22.6 45 47.9 45H346.9c25.3 0 46.3-19.7 47.9-45L416 128z"/>');
                break;
            case 'downvote':
                title = 'Downvote';
                svg = UIToolkit.svg(320, 512, '<path fill="currentColor" d="M318 334.5c3.8 8.8 2 19-4.6 26l-136 144c-4.5 4.8-10.8 7.5-17.4 7.5s-12.9-2.7-17.4-7.5l-136-144c-6.6-7-8.4-17.2-4.6-26S14.4 320 24 320h88l0-288c0-17.7 14.3-32 32-32h32c17.7 0 32 14.3 32 32l0 288h88c9.6 0 18.2 5.7 22 14.5z"/>');
                break;
            case 'edit':
                title = 'Edit';
                svg = UIToolkit.svg(512, 512, '<path fill="currentColor" d="M362.7 19.3L314.3 67.7 444.3 197.7l48.4-48.4c25-25 25-65.5 0-90.5L453.3 19.3c-25-25-65.5-25-90.5 0zm-71 71L58.6 323.5c-10.4 10.4-18 23.3-22.2 37.4L1 481.2C-1.5 489.7 .8 498.8 7 505s15.3 8.5 23.7 6.1l120.3-35.4c14.1-4.2 27-11.8 37.4-22.2L421.7 220.3 291.7 90.3z"/>');
                break;
            case 'reply':
                title = 'Reply';
                svg = UIToolkit.svg(512, 512, '<path fill="currentColor" d="M205 34.8c11.5 5.1 19 16.6 19 29.2v64H336c97.2 0 176 78.8 176 176c0 113.3-81.5 163.9-100.2 174.1c-2.5 1.4-5.3 1.9-8.1 1.9c-10.9 0-19.7-8.9-19.7-19.7c0-7.5 4.3-14.4 9.8-19.5c9.4-8.8 22.2-26.4 22.2-56.7c0-53-43-96-96-96H224v64c0 12.6-7.4 24.1-19 29.2s-25 3-34.4-5.4l-160-144C3.9 225.7 0 217.1 0 208s3.9-17.7 10.6-23.8l160-144c9.4-8.5 22.9-10.6 34.4-5.4z"/>');
                break;
            case 'sticky':
                title = 'Sticky';
                svg = UIToolkit.svg(576, 512, '<path fill="currentColor" d="M316.9 18C311.6 7 300.4 0 288.1 0s-23.4 7-28.8 18L195 150.3 51.4 171.5c-12 1.8-22 10.2-25.7 21.7s-.7 24.2 7.9 32.7L137.8 329 113.2 474.7c-2 12 3 24.2 12.9 31.3s23 8 33.8 2.3l128.3-68.5 128.3 68.5c10.8 5.7 23.9 4.9 33.8-2.3s14.9-19.3 12.9-31.3L438.5 329 542.7 225.9c8.6-8.5 11.7-21.2 7.9-32.7s-13.7-19.9-25.7-21.7L381.2 150.3 316.9 18z"/>');
                break;
            case 'upvote':
                title = 'Upvote';
                svg = UIToolkit.svg(320, 512, '<path fill="currentColor" d="M318 177.5c3.8-8.8 2-19-4.6-26l-136-144C172.9 2.7 166.6 0 160 0s-12.9 2.7-17.4 7.5l-136 144c-6.6 7-8.4 17.2-4.6 26S14.4 192 24 192h88l0 288c0 17.7 14.3 32 32 32h32c17.7 0 32-14.3 32-32l0-288h88c9.6 0 18.2-5.7 22-14.5z"/>');
                break;
        }
        return UIToolkit.button(svg, onClick, 'option-button', cls).attr({title});
    }
}
