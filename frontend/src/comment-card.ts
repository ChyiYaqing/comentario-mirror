import { Wrap } from './element-wrap';
import { Comment, CommenterMap, CommentsGroupedByHex, sortingProps, SortPolicy } from './models';
import { UIToolkit } from './ui-toolkit';
import { Utils } from './utils';

// eslint-disable-next-line no-use-before-define
export type CommentCardMap = { [k: string]: CommentCard };

// eslint-disable-next-line no-use-before-define
export type CommentCardEventHandler = (c: CommentCard) => void;

/**
 * Context for rendering comment trees.
 */
export interface CommentRenderingContext {
    /** Base CDN URL. */
    readonly cdn: string;
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
    /** Card map populated during rendering. */
    readonly cardMap: CommentCardMap;

    // Events
    readonly onApprove: CommentCardEventHandler;
    readonly onDelete: CommentCardEventHandler;
    readonly onEdit: CommentCardEventHandler;
    readonly onReply: CommentCardEventHandler;
    readonly onSticky: CommentCardEventHandler;
    readonly onVoteDown: CommentCardEventHandler;
    readonly onVoteUp: CommentCardEventHandler;
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
            // eslint-disable-next-line no-use-before-define
            .map(c => new CommentCard(c).render(ctx));

        // If there's any cards, return it wrapped in a .body
        return cards?.length ? UIToolkit.div('body').append(...cards) : null;
    }
}

/**
 * Comment card represents an individual comment in the UI.
 */
export class CommentCard {

    btnCollapse: Wrap<HTMLButtonElement>;
    btnSticky: Wrap<HTMLButtonElement>;
    children?: Wrap<HTMLDivElement>;
    _collapsed = false;

    constructor(
        readonly comment: Comment,
    ) {}

    /**
     * Set the current card's children collapsed state.
     */
    set collapsed(c: boolean) {
        this._collapsed = c;

        // Set children visibility
        this.children?.classes(c && 'hidden').noClasses(!c && 'hidden');

        // Set the button appearance
        this.btnCollapse
            ?.noClasses(c && 'option-collapse', !c && 'option-uncollapse')
            .classes(!c && 'option-collapse', c && 'option-uncollapse');
    }

    set sticky(b: boolean) {
        this.btnSticky
            ?.noClasses(!b && 'option-unsticky', b && 'option-sticky')
            .classes(b && 'option-unsticky', !b && 'option-sticky');
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

        // Store the card in the context
        ctx.cardMap[hex] = this;

        // Render a card
        return UIToolkit.div('card', ctx.isModerator && this.comment.state !== 'approved' && 'dark-card', `border-${idxColor}`)
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
                        Wrap.new(commLink ? 'a' : 'div')
                            //TODO .id(IDS.name + hex)
                            .inner(this.comment.deleted ? '[deleted]' : commenter.name)
                            .classes(
                                'name',
                                commenter.isModerator && 'moderator',
                                this.comment.state === 'flagged' && 'flagged')
                            .attr({href: commLink, rel: commLink && 'nofollow noopener noreferrer'}),
                        // Subtitle
                        UIToolkit.div('subtitle')
                            .append(
                                // Score
                                UIToolkit.div('score')/* TODO .id(IDS.score + hex)*/.inner(Utils.score(this.comment.score)),
                                // Time ago
                                UIToolkit.div('timeago')
                                    .inner(Utils.timeAgo(ctx.curTimeMs, this.comment.creationMs))
                                    .attr({title: this.comment.creationDate.toString()}))),
                // Card contents
                UIToolkit.div()
                    .append(
                        UIToolkit.div('body')
                            //TODO .id(IDS.body + hex)
                            .append(UIToolkit.div()/*TODO .id(IDS.text + hex)*/.html(this.comment.html)),
                        this.children));
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
            Wrap.new('button')
                .classes('option-button', 'option-sticky')
                .attr({title: 'This comment has been stickied', type: 'button', disabled: 'true'})
                .appendTo(options);
        }

        // Approve button
        if (ctx.isModerator && this.comment.state !== 'approved') {
            Wrap.new('button')
                // TODO .id(IDS.approve + hex)
                .classes('option-button', 'option-approve')
                .attr({type: 'button', title: 'Approve'})
                // TODO .click(() => this.commentApprove(hex))
                .appendTo(options);
        }

        // Remove button
        if (!this.comment.deleted && (ctx.isModerator || this.comment.commenterHex === ctx.selfHex)) {
            Wrap.new('button')
                .classes('option-button', 'option-remove')
                .attr({type: 'button', title: 'Remove'})
                // TODO .click(btn => this.commentDelete(btn, hex))
                .appendTo(options);
        }

        // Sticky toggle button (for moderator and a top-level comments only)
        if (!this.comment.deleted && ctx.isModerator && parentHex === 'root') {
            this.btnSticky = Wrap.new('button')
                .classes('option-button', isSticky ? 'option-unsticky' : 'option-sticky')
                .attr({title: isSticky ? 'Unsticky' : 'Sticky', type: 'button'})
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

        // Upvote / Downvote buttons
        // TODO if (!this.comment.deleted) {
        // TODO     this.updateUpDownAction(
        // TODO         Wrap.new('button')
        // TODO             .id(IDS.upvote + hex)
        // TODO             .classes('option-button', 'option-upvote', ctx.isAuthenticated && this.comment.direction > 0 && 'upvoted')
        // TODO             .attr({type: 'button', title: 'Upvote'})
        // TODO             .appendTo(options),
        // TODO         Wrap.new('button')
        // TODO             .id(IDS.downvote + hex)
        // TODO             .classes('option-button', 'option-downvote', ctx.isAuthenticated && this.comment.direction < 0 && 'downvoted')
        // TODO             .attr({type: 'button', title: 'Downvote'})
        // TODO             .appendTo(options),
        // TODO         hex,
        // TODO         this.comment.direction);
        // TODO }

        // Collapse button, if there are any children
        if (this.children?.ok) {
            this.btnCollapse = Wrap.new('button')
                .classes('option-button', 'option-collapse')
                .attr({type: 'button', title: 'Collapse children'})
                .click(() => this.collapsed = !this._collapsed)
                .appendTo(options);
        }
        return options;
    }
}
