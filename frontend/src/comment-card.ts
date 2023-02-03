import { Wrap } from './element-wrap';
import { Comment, CommenterMap, CommentsGroupedByHex, sortingProps, SortPolicy } from './models';
import { UIToolkit } from './ui-toolkit';
import { Utils } from './utils';
import { ConfirmDialog } from './confirm-dialog';

export type CommentCardEventHandler = (c: CommentCard) => void;
export type CommentCardVoteEventHandler = (c: CommentCard, direction: -1 | 0 | 1) => void;

/**
 * Context for rendering comment trees.
 */
export interface CommentRenderingContext {
    /** Base CDN URL. */
    readonly cdn: string;
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
    children: Wrap<HTMLDivElement>;

    private eName?: Wrap<HTMLDivElement | HTMLAnchorElement>;
    private eScore?: Wrap<HTMLDivElement>;
    private eHeader: Wrap<HTMLDivElement>;
    private eBody: Wrap<HTMLDivElement>;
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
     * Insert the given card as the first child comment card.
     * @param card Card to insert.
     */
    prependCard(card: CommentCard) {
        this.children.prepend(card);
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
            this.children.remove();

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
            ?.inner(c.score.toString())
            .setClasses(c.score > 0, 'score-upvoted').setClasses(c.score < 0, 'score-downvoted');
        this.btnUpvote?.setClasses(c.direction > 0, 'upvoted');
        this.btnDownvote?.setClasses(c.direction < 0, 'downvoted');

        // Collapsed
        this.btnCollapse
            ?.attr({title: this.collapsed ? 'Expand children' : 'Collapse children'})
            .setClasses(this.collapsed, 'option-uncollapse').setClasses(!this.collapsed, 'option-collapse');

        // Approved
        const flagged = this.flagged;
        this.setClasses(flagged, 'dark-card');
        this.eName?.setClasses(flagged, 'flagged');
        if (!flagged && this.btnApprove) {
            // Remove the Approve button if the comment is approved
            this.btnApprove.remove();
            this.btnApprove = null;
        }

        // Moderation notice
        let mn: string;
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
                this.eModNotice = UIToolkit.div('moderation-notice').appendTo(this.eHeader);
            }
            this.eModNotice.inner(mn);

        // No moderation notice
        } else if (this.eModNotice) {
            this.eModNotice.remove();
            this.eModNotice = null;
        }
    }

    /**
     * Update the current comment's text.
     */
    updateText() {
        this.eBody.html(this.comment.html);
    }

    /**
     * Render the content of the card.
     * @private
     */
    private render(ctx: CommentRenderingContext) {
        const hex = this.comment.commentHex;
        const commenter = ctx.commenters[this.comment.commenterHex];

        // Figure out if the commenter has a profile link
        const commLink = !commenter.link || commenter.link === 'undefined' || commenter.link === 'https://undefined' ? undefined : commenter.link;

        // Pick a color for the commenter
        const idxColor = Utils.colourIndex(`${this.comment.commenterHex}-${commenter.name}`);

        // Render children
        this.children = UIToolkit.div('card-children').append(...new CommentTree().render(ctx, hex));

        // Render a card
        this.id(`card-${hex}`) // ID for scrolling to
            .classes('card', `border-${idxColor}`)
            .append(
                // Card header
                this.eHeader = UIToolkit.div('card-header')
                    .append(
                        // Options toolbar
                        this.commentOptionsBar(ctx, hex, this.comment.parentHex),
                        // Avatar
                        commenter.photo === 'undefined' ?
                            UIToolkit.div('avatar', `bg-${idxColor}`)
                                .html(this.comment.commenterHex === 'anonymous' ? '?' : commenter.name[0].toUpperCase()) :
                            Wrap.new('img')
                                .classes('avatar-img')
                                .attr({
                                    src: `${ctx.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                                    alt: '',
                                }),
                        // Name
                        this.eName = Wrap.new(commLink ? 'a' : 'div')
                            .inner(this.comment.deleted ? '[deleted]' : commenter.name)
                            .classes('name', commenter.isModerator && 'moderator')
                            .attr({href: commLink, rel: commLink && 'nofollow noopener noreferrer'}),
                        // Subtitle
                        UIToolkit.div('subtitle')
                            // Time ago
                            .append(
                                UIToolkit.div('timeago')
                                    .inner(Utils.timeAgo(ctx.curTimeMs, this.comment.creationMs))
                                    .attr({title: this.comment.creationDate.toString()}))),
                // Card body
                this.eBody = UIToolkit.div('card-body'),
                // Children (if any)
                this.children);
    }

    /**
     * Return a wrapped options toolbar for a comment.
     * @private
     */
    private commentOptionsBar(ctx: CommentRenderingContext, hex: string, parentHex: string): Wrap<HTMLDivElement> {
        const options = UIToolkit.div('options');

        // Sticky comment indicator (for non-moderator only)
        const isSticky = ctx.stickyHex === hex;
        if (!this.comment.deleted && !ctx.isModerator && isSticky) {
            this.btnSticky = Wrap.new('button')
                .classes('option-button')
                .attr({type: 'button', disabled: 'true'})
                .appendTo(options);
        }

        // Approve button
        if (ctx.isModerator && this.comment.state !== 'approved') {
            this.btnApprove = Wrap.new('button')
                .classes('option-button', 'option-approve')
                .attr({type: 'button', title: 'Approve'})
                .click(() => ctx.onApprove(this))
                .appendTo(options);
        }

        // Delete button
        if (!this.comment.deleted && (ctx.isModerator || this.comment.commenterHex === ctx.selfHex)) {
            this.btnDelete = Wrap.new('button')
                .classes('option-button', 'option-remove')
                .attr({type: 'button', title: 'Remove'})
                .click(btn => this.deleteComment(btn, ctx))
                .appendTo(options);
        }

        // Sticky toggle button (for moderator and top-level comments only)
        if (!this.comment.deleted && ctx.isModerator && parentHex === 'root') {
            this.btnSticky = Wrap.new('button')
                .classes('option-button')
                .attr({type: 'button'})
                .click(() => ctx.onSticky(this))
                .appendTo(options);
        }

        // Own comment: Edit button
        if (this.comment.commenterHex === ctx.selfHex) {
            this.btnEdit = Wrap.new('button')
                .classes('option-button', 'option-edit')
                .attr({type: 'button', title: 'Edit'})
                .click(() => ctx.onEdit(this))
                .appendTo(options);

        // Someone other's comment: Reply button
        } else if (!ctx.isLocked && !this.comment.deleted) {
            this.btnReply = Wrap.new('button')
                .classes('option-button', 'option-reply')
                .attr({type: 'button', title: 'Reply'})
                .click(() => ctx.onReply(this))
                .appendTo(options);
        }

        // Collapse button, if there are any children
        if (this.children.hasChildren) {
            this.btnCollapse = Wrap.new('button')
                .classes('option-button')
                .attr({type: 'button'})
                .click(() => this.collapse(!this.collapsed))
                .appendTo(options);
        }

        // Upvote / Downvote buttons and the score
        if (!this.comment.deleted) {
            options.append(
                this.btnUpvote = Wrap.new('button')
                    .classes('option-button', 'option-upvote')
                    .attr({type: 'button', title: 'Upvote'})
                    .click(() => ctx.onVote(this, this.comment.direction > 0 ? 0 : 1)),
                this.eScore = UIToolkit.div('score').attr({title: 'Comment score'}),
                this.btnDownvote = Wrap.new('button')
                    .classes('option-button', 'option-downvote')
                    .attr({type: 'button', title: 'Downvote'})
                    .click(() => ctx.onVote(this, this.comment.direction < 0 ? 0 : -1)));
        }

        // Update the sticky button, if any (the sticky status can only be changed after a full tree reload)
        this.btnSticky
            ?.classes(isSticky ? 'option-unsticky' : 'option-sticky')
            .attr({title: isSticky ? (ctx.isModerator ? 'Unsticky' : 'This comment has been stickied') : 'Sticky'});
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
}
