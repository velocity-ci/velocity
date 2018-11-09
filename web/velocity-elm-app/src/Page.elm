module Page exposing (Page(..), view, viewErrors)

import Api exposing (Cred)
import Asset
import Browser exposing (Document)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html)
import Route exposing (Route)
import Session exposing (Session)
import Username exposing (Username)
import Viewer exposing (Viewer)


maxWidth : Int
maxWidth =
    1280


{-| Determines which navbar link (if any) will be rendered as active.

Note that we don't enumerate every page here, because the navbar doesn't
have links for every page. Anything that's not part of the navbar falls
under Other.

-}
type Page
    = Other
    | Home
    | Login
    | Register
    | Settings
    | Profile Username
    | NewArticle


{-| Take a page's Html and frames it with a header and footer.

The caller provides the current user, so we can display in either
"signed in" (rendering username) or "signed out" mode.

isLoading is for determining whether we should show a loading spinner
in the header. (This comes up during slow page transitions.)

-}
view :
    Maybe Viewer
    -> Page
    -> { title : String, content : Element msg }
    -> (msg -> msg2)
    -> { title : String, body : List (Html msg2) }
view maybeViewer page { title, content } toMsg =
    { title = title ++ " - Conduit"
    , body =
        [ Element.layout
            [ Font.family
                [ Font.typeface "Roboto"
                , Font.sansSerif
                ]
            ]
            (Element.column
                [ width fill
                , height fill
                ]
                [ viewHeader page maybeViewer
                , viewBody content toMsg
                , viewFooter
                ]
            )
        ]
    }


viewBody : Element msg -> (msg -> msg2) -> Element msg2
viewBody content toMsg =
    row
        [ width (fill |> maximum maxWidth)
        , height fill
        , centerX
        , paddingXY 20 0
        ]
        [ Element.map toMsg content ]


viewHeader : Page -> Maybe Viewer -> Element msg
viewHeader page maybeViewer =
    row
        [ width fill
        , height (px 55)
        , Border.shadow
            { offset = ( 0, 2 )
            , size = 2
            , blur = 2
            , color = rgba255 245 245 245 1
            }
        , Background.color (rgba255 245 245 245 0.5)
        ]
        [ row
            [ width (fill |> maximum maxWidth)
            , centerX
            , paddingXY 20 0
            ]
            [ el [ alignLeft ] viewBrand
            , row
                [ centerY
                , alignRight
                , spacing 10
                ]
              <|
                viewMenu page maybeViewer
            ]
        ]


viewBrand : Element msg
viewBrand =
    el
        [ Font.color (rgba255 92 184 92 1)
        , Font.heavy
        , Font.size 28
        , Font.letterSpacing -1
        , Font.family
            [ Font.typeface "titillium web"
            , Font.sansSerif
            ]
        ]
        (text "Velocity")


viewMenu : Page -> Maybe Viewer -> List (Element msg)
viewMenu page maybeViewer =
    let
        linkTo =
            navbarLink page
    in
    case maybeViewer of
        Just viewer ->
            [ linkTo (Route.Home Nothing)
                (el
                    [ width (px 30)
                    , height (px 30)
                    , Border.rounded 180
                    , Background.image (Asset.src Asset.defaultAvatar)
                    ]
                    (text "")
                )
            ]

        Nothing ->
            [ linkTo Route.Login (text "Sign in")
            ]


viewFooter : Element msg
viewFooter =
    column
        [ width fill
        , height (px 50)
        , Border.shadow
            { offset = ( 2, 2 )
            , size = 2
            , blur = 2
            , color = rgba255 245 245 245 1
            }
        , Background.color (rgba255 245 245 245 0.5)
        ]
        [ Element.el
            [ centerY
            , centerX
            , width (fill |> maximum maxWidth)
            ]
            (text "")
        ]


navbarLink : Page -> Route -> Element msg -> Element msg
navbarLink page route linkContent =
    Route.link
        (if isActive page route then
            [ Font.color (rgba255 0 0 0 0.8) ]

         else
            [ Font.color (rgba255 0 0 0 0.3)
            , mouseOver [ Font.color (rgba255 0 0 0 0.5) ]
            ]
        )
        linkContent
        route


isActive : Page -> Route -> Bool
isActive page route =
    case ( page, route ) of
        ( Home, Route.Home _ ) ->
            True

        ( Login, Route.Login ) ->
            True

        _ ->
            False


{-| Render dismissable errors. We use this all over the place!
-}
viewErrors : msg -> List String -> Element msg
viewErrors dismissErrors errors =
    if List.isEmpty errors then
        text ""

    else
        row [] <|
            List.map (\error -> paragraph [] [ text error ]) errors
                ++ [ Input.button [] { onPress = Just dismissErrors, label = text "Dismiss errors" } ]
