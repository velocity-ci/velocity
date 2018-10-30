module Views.Style exposing (textOverflowMixin)

import Css exposing (..)


textOverflowMixin : Style
textOverflowMixin =
    Css.batch
        [ whiteSpace noWrap
        , overflow hidden
        , textOverflow ellipsis
        ]
