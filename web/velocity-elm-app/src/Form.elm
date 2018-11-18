module Form exposing (buttonAttrs)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Palette



-- Attrs and decorations


buttonAttrs : List (Element.Attribute msg)
buttonAttrs =
    [ width fill
    , height (px 45)
    , Border.width 1
    , Border.rounded 5
    , Border.color Palette.primary4
    , Font.color Palette.white
    , Background.color Palette.primary3
    , padding 11
    , Font.size 16
    , alignBottom
    , pointer
    , mouseOver
        [ Background.color Palette.primary4
        , Font.color Palette.white
        ]
    ]
