package types

/*
This package contains types that are used across multiple pipeviz subcomponents.
To avoid import cycles, these types would either have to live in the earliest
component package where they appear, or live in a separate package that all
relevant components can safely import.

A separate package - this one - is used because it helps make clearer what the
shared pieces are, which is boon for architectural clarity.

Most of what's here is interfaces. The concrete real types are mostly inert,
or have straightforward helper methods.
*/
