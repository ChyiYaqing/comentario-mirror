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
    render(ctx: CommentRenderingContext, parentHex: string): Wrap<any> | null {
        // Fetch comments that have the given parentHex
        const comments = ctx.parentMap[parentHex];

        // Apply the chosen sorting, always keeping the sticky comment on top
        comments?.sort((a, b) =>
            !a.deleted && a.commentHex === ctx.stickyHex ?
                -Infinity :
                !b.deleted && b.commentHex === ctx.stickyHex ?
                    Infinity :
                    sortingProps[ctx.sortPolicy].comparator(a, b));

        // Render child comments, if any
        const cards = comments
            // Filter out deleted comment, if they're to be hidden
            ?.filter(c => !ctx.hideDeleted || !c.deleted)
            // Render a comment card
            // eslint-disable-next-line @typescript-eslint/no-use-before-define
            .map(c => new CommentCard(c).render(ctx));

        // If there's any cards, return it wrapped in a .body
        return cards?.length ? UIToolkit.div('body').append(...cards) : null;
    }
}

/**
 * Comment card represents an individual comment in the UI.
 */
export class CommentCard {

    private children?: Wrap<HTMLDivElement>;
    private eCard?: Wrap<HTMLDivElement>;
    private eName?: Wrap<HTMLDivElement | HTMLAnchorElement>;
    private eScore?: Wrap<HTMLDivElement>;
    private eText?: Wrap<HTMLDivElement>;
    private btnApprove: Wrap<HTMLButtonElement>;
    private btnCollapse: Wrap<HTMLButtonElement>;
    private btnDelete: Wrap<HTMLButtonElement>;
    private btnDownvote: Wrap<HTMLButtonElement>;
    private btnUpvote: Wrap<HTMLButtonElement>;
    private collapsed = false;

    constructor(
        readonly comment: Comment,
    ) {}

    /**
     * Current comment's flagged state.
     */
    get flagged(): boolean {
        return this.comment.state === 'flagged';
    }

    render(ctx: CommentRenderingContext): Wrap<HTMLDivElement> {
        const hex = this.comment.commentHex;
        const commenter = ctx.commenters[this.comment.commenterHex];

        // Figure out if the commenter has a profile link
        const commLink = !commenter.link || commenter.link === 'undefined' || commenter.link === 'https://undefined' ? undefined : commenter.link;

        // Pick a color for the commenter
        const idxColor = Utils.colourIndex(`${this.comment.commenterHex}-${commenter.name}`);

        // Render children
        this.children = new CommentTree().render(ctx, hex);

        // Render a card
        this.eName = Wrap.new(commLink ? 'a' : 'div')
            .inner(this.comment.deleted ? '[deleted]' : commenter.name)
            .classes('name', commenter.isModerator && 'moderator')
            .attr({href: commLink, rel: commLink && 'nofollow noopener noreferrer'});
        this.eCard = UIToolkit.div('card', `border-${idxColor}`)
            .append(
                // Card header
                UIToolkit.div('header')
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
                        this.eName,
                        // Subtitle
                        UIToolkit.div('subtitle')
                            // Time ago
                            .append(
                                UIToolkit.div('timeago')
                                    .inner(Utils.timeAgo(ctx.curTimeMs, this.comment.creationMs))
                                    .attr({title: this.comment.creationDate.toString()}))),
                // Card contents
                UIToolkit.div()
                    .append(
                        UIToolkit.div('body')
                            //TODO .id(IDS.body + hex)
                            .append(this.eText = UIToolkit.div().html(this.comment.html)),
                        this.children));

        // Update the card controls
        this.update();
        return this.eCard;
    }

    /**
     * Update comment controls according to the related comment's properties.
     */
    update() {
        const c = this.comment;
        this.eScore
            ?.inner(c.score.toString())
            .setClasses(c.score > 0, 'score-upvoted').setClasses(c.score < 0, 'score-downvoted');
        this.btnUpvote?.setClasses(c.direction > 0, 'upvoted');
        this.btnDownvote?.setClasses(c.direction < 0, 'downvoted');

        // Collapsed
        this.btnCollapse
            ?.attr({title: this.collapsed ? 'Expand children' : 'Collapse children'})
            .setClasses(this.collapsed, 'option-uncollapse').setClasses(!this.collapsed, 'option-collapse');

        // Deleted
        if (c.deleted) {
            this.eText?.inner('[deleted]');
            // TODO also remove all option buttons, except Collapse, and (?) child comments
        }

        // Approved
        const flagged = this.flagged;
        this.eCard?.setClasses(flagged, 'dark-card');
        this.eName?.setClasses(flagged, 'flagged');
        if (!flagged && this.btnApprove) {
            // Remove the Approve button if the comment is approved
            this.btnApprove.remove();
            this.btnApprove = null;
        }
    }

    /**
     * Return a wrapped options toolbar for a comment.
     * @private
     */
    private commentOptionsBar(ctx: CommentRenderingContext, hex: string, parentHex: string): Wrap<HTMLDivElement> {
        const options = UIToolkit.div('options');
        let btnSticky: Wrap<HTMLButtonElement>;

        // Sticky comment indicator (for non-moderator only)
        const isSticky = ctx.stickyHex === hex;
        if (!this.comment.deleted && !ctx.isModerator && isSticky) {
            btnSticky = Wrap.new('button')
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
            btnSticky = Wrap.new('button')
                .classes('option-button')
                .attr({type: 'button'})
                .click(() => ctx.onSticky(this))
                .appendTo(options);
        }

        // Own comment: Edit button
        if (this.comment.commenterHex === ctx.selfHex) {
            Wrap.new('button')
                // TODO .id(IDS.edit + hex)
                .classes('option-button', 'option-edit')
                .attr({type: 'button', title: 'Edit'})
                // TODO .click(() => this.startEditing(hex))
                .appendTo(options);

        // Someone other's comment: Reply button
        } else if (!this.comment.deleted) {
            Wrap.new('button')
                // TODO .id(IDS.reply + hex)
                .classes('option-button', 'option-reply')
                .attr({type: 'button', title: 'Reply'})
                // TODO .click(() => this.replyShow(hex))
                .appendTo(options);
        }

        // Collapse button, if there are any children
        if (this.children?.ok) {
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
        btnSticky
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
