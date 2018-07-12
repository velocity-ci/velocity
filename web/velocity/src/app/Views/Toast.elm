module Views.Toast exposing (config, genericToast)

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (style, css, class, classList)
import Html.Attributes as UnstyledAttribute
import Html.Styled as Styled exposing (..)
import Css exposing (..)
import Data.Event as Event exposing (Event)
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
    [ UnstyledAttribute.style
        [ ( "position", "fixed" )
        , ( "top", "0" )
        , ( "right", "0" )
        , ( "width", "100%" )
        , ( "max-width", "300px" )
        , ( "list-style-type", "none" )
        , ( "padding", "0" )
        , ( "margin", "0" )
        ]
    ]


itemAttrs : List (Html.Attribute msg)
itemAttrs =
    [ UnstyledAttribute.style
        [ ( "margin", "1em 1em 0 1em" )
        , ( "max-height", "100px" )
        , ( "transition", "max-height 0.6s, margin-top 0.6s" )
        ]
    ]


transitionInAttrs : List (Html.Attribute msg)
transitionInAttrs =
    [ UnstyledAttribute.class "animated bounceInRight"
    ]


transitionOutAttrs : List (Html.Attribute msg)
transitionOutAttrs =
    [ UnstyledAttribute.class "animated fadeOutRightBig"
    , UnstyledAttribute.style
        [ ( "max-height", "0" )
        , ( "margin-top", "0" )
        ]
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
            [ if (String.isEmpty message) then
                text ""
              else
                text message
            ]
        ]
        |> toUnstyled
