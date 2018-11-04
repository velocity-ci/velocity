module Page exposing (Page(..), view, viewErrors)

import Api exposing (Cred)
import Browser exposing (Document)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Element.Input as Input
import Route exposing (Route)
import Session exposing (Session)
import Username exposing (Username)
import Viewer exposing (Viewer)


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
view : Maybe Viewer -> Page -> { title : String, content : Element msg } -> { title : String, body : Element msg }
view maybeViewer page { title, content } =
    { title = title ++ " - Conduit"
    , body = Element.column [ width fill, height fill ] (viewHeader page maybeViewer :: content :: [ viewFooter ])
    }


viewHeader : Page -> Maybe Viewer -> Element msg
viewHeader page maybeViewer =
    row
        [ width fill
        , height (px 55)
        , padding 20
        , Border.shadow
            { offset = ( 0, 2 )
            , size = 2
            , blur = 2
            , color = rgba255 245 245 245 1
            }
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
            [ linkTo Route.Home (text "Home")
            , linkTo Route.Logout (text "Sign out")
            ]

        Nothing ->
            [ linkTo Route.Login (text "Sign in")
            ]


viewFooter : Element msg
viewFooter =
    column
        [ width fill
        , height (px 50)
        , Background.color (rgb 0 0.5 0)
        , Border.color (rgb 0 0.7 0)
        ]
        [ Element.el [ centerY ] (text "Footer") ]


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
        ( Home, Route.Home ) ->
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
