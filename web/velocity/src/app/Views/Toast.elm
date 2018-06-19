module Views.Toast exposing (config, genericToast)

import Html.Attributes exposing (..)
import Html exposing (..)
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
    [ style
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
    [ style
        [ ( "margin", "1em 1em 0 1em" )
        , ( "max-height", "100px" )
        , ( "transition", "max-height 0.6s, margin-top 0.6s" )
        ]
    ]


transitionInAttrs : List (Html.Attribute msg)
transitionInAttrs =
    [ class "animated bounceInRight"
    ]


transitionOutAttrs : List (Html.Attribute msg)
transitionOutAttrs =
    [ class "animated fadeOutRightBig"
    , style
        [ ( "max-height", "0" )
        , ( "margin-top", "0" )
        ]
    ]


genericToast : String -> String -> String -> Html msg
genericToast variantClass title message =
    div
        [ class "toasty-container", class variantClass ]
        [ h1 [ class "toasty-title" ] [ text title ]
        , if (String.isEmpty message) then
            text ""
          else
            p [ class "toasty-message" ] [ text message ]
        ]
