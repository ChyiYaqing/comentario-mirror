import { Comment, Commenter, Email, SortPolicy } from './models';

export interface ApiResponseBase {
    success: boolean;
    message: string;
}

export interface ApiSelfResponse extends ApiResponseBase {
    commenter: Commenter;
    email:     Email;
}

export interface ApiCommentListResponse extends ApiResponseBase {
    requireIdentification: boolean;
    isModerator:           boolean;
    isFrozen:              boolean;
    attributes:            any;
    comments:              Comment[];
    commenters:            Commenter[];
    configuredOauths:      { [k: string]: boolean };
    defaultSortPolicy:     SortPolicy;
}

export interface ApiCommentNewResponse extends ApiResponseBase {
    state:      'unapproved' | 'flagged';
    commentHex: string;
    html:       string;
}

export interface ApiCommentEditResponse extends ApiResponseBase {
    state:      'unapproved' | 'flagged';
    commentHex: string;
    html:       string;
}

export interface ApiCommenterTokenNewResponse extends ApiResponseBase {
    commenterToken: string;
}

export interface ApiCommenterLoginResponse extends ApiResponseBase {
    commenterToken: string;
    commenter:      Commenter;
    email:          Email;
}
