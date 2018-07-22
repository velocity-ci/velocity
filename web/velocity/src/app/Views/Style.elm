module Views.Style exposing (..)

import Css exposing (..)


textOverflowMixin : Style
textOverflowMixin =
    Css.batch
        [ whiteSpace noWrap
        , overflow hidden
        , textOverflow ellipsis
        ]
