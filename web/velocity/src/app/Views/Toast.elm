module Views.Toast exposing (config, genericToast)

import Css exposing (..)
import Data.Event as Event exposing (Event)
import Html exposing (Html)
import Html.Attributes as UnstyledAttribute
import Html.Styled as Styled exposing (..)
import Html.Styled.Attributes as Attributes exposing (class, classList, css, style)
import Toasty


{-| Default theme configuration.
-}
config : Toasty.Config msg
config =
    Toasty.config
        |> Toasty.transitionOutDuration 7500
        |> Toasty.transitionOutAttrs transitionOutAttrs
        |> Toasty.transitionInAttrs transitionInAttrs
        |> Toasty.containerAttrs containerAttrs
        |> Toasty.itemAttrs itemAttrs
        |> Toasty.delay 7500


containerAttrs : List (Html.Attribute msg)
containerAttrs =
    [ UnstyledAttribute.style "position" "fixed"
    , UnstyledAttribute.style "top" "0"
    , UnstyledAttribute.style "right" "0"
    , UnstyledAttribute.style "width" "100%"
    , UnstyledAttribute.style "max-width" "300px"
    , UnstyledAttribute.style "list-style-type" "none"
    , UnstyledAttribute.style "padding" "0"
    , UnstyledAttribute.style "margin" "0"
    ]


itemAttrs : List (Html.Attribute msg)
itemAttrs =
    [ UnstyledAttribute.style "margin" "1em 1em 0 1em"
    , UnstyledAttribute.style "max-height" "100px"
    , UnstyledAttribute.style "transition" "max-height 0.6s, margin-top 0.6s"
    ]


transitionInAttrs : List (Html.Attribute msg)
transitionInAttrs =
    [ UnstyledAttribute.class "animated bounceInRight"
    ]


transitionOutAttrs : List (Html.Attribute msg)
transitionOutAttrs =
    [ UnstyledAttribute.class "animated fadeOutRightBig"
    , UnstyledAttribute.style "max-height" "0"
    , UnstyledAttribute.style "margin-top" "0"
    ]


genericToast : String -> String -> String -> Html.Html msg
genericToast variantClass title message =
    div
        [ css
            [ padding (Css.em 1)
            , borderRadius (px 5)
            , cursor pointer
            , boxShadow5 (px 0) (px 5) (px 5) (px -5) (rgba 0 0 0 0.5)
            , color (hex "ffffff")
            , fontSize (px 13)
            ]
        , class variantClass
        ]
        [ h1
            [ css
                [ fontSize (Css.em 1)
                , margin (px 0)
                ]
            ]
            [ text title ]
        , p
            [ css
                [ fontSize (Css.em 0.9)
                , marginTop (Css.em 0.5)
                , marginBottom (Css.em 0)
                ]
            ]
            [ if String.isEmpty message then
                text ""
              else
                text message
            ]
        ]
        |> toUnstyled
