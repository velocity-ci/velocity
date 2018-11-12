module Page exposing (Header, Page(..), initHeader, view, viewErrors)

import Api exposing (Cred)
import Asset
import Browser exposing (Document)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html)
import Palette
import Route exposing (Route)
import Username exposing (Username)
import Viewer exposing (Viewer)



-- Header


type Header
    = Header Internals


type alias Internals =
    { userMenuOpen : Bool }


initHeader : Header
initHeader =
    Header (Internals False)



-- Max width of frame


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
type alias Config msg pageMsg =
    { viewer : Maybe Viewer
    , page : Page
    , title : String
    , content : Element msg
    , toMsg : msg -> pageMsg
    , header : Header
    , updateHeader : Header -> pageMsg
    }


view : Config subMsg msg -> { title : String, body : List (Html msg) }
view { viewer, page, title, content, toMsg, header, updateHeader } =
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
                [ viewHeader page viewer updateHeader header
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


viewHeader : Page -> Maybe Viewer -> (Header -> msg) -> Header -> Element msg
viewHeader page maybeViewer headerMsg header =
    row
        [ width fill
        , height (px 55)
        , Border.shadow
            { offset = ( 0, 2 )
            , size = 2
            , blur = 2
            , color = Palette.neutral6
            }
        , Background.color Palette.neutral7
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
                viewMenu page maybeViewer headerMsg header
            ]
        ]


viewBrand : Element msg
viewBrand =
    el
        [ Font.color Palette.primary4
        , Font.heavy
        , Font.size 28
        , Font.letterSpacing -1
        , Font.family
            [ Font.typeface "titillium web"
            , Font.sansSerif
            ]
        ]
        (text "Velocity")


viewMenu : Page -> Maybe Viewer -> (Header -> msg) -> Header -> List (Element msg)
viewMenu page maybeViewer headerMsg (Header { userMenuOpen }) =
    let
        linkTo =
            navbarLink page

        dropdownMenu =
            if userMenuOpen then
                column
                    [ Background.color Palette.neutral7
                    , Border.color Palette.neutral4
                    , Border.width 1
                    , Border.rounded 7
                    , moveRight -170
                    , width (px 200)
                    ]
                    [ row
                        [ width fill
                        , padding 10
                        , spacingXY 10 0
                        ]
                        [ el
                            [ width (px 45)
                            , height (px 45)
                            , Border.rounded 90
                            , Background.image (Asset.src Asset.defaultAvatar)
                            ]
                            (text "")
                        , column [ width fill ]
                            [ paragraph [ Font.size 15, Font.color Palette.primary2 ] [ text "Signed in as" ]
                            , paragraph [ Font.size 18, Font.heavy, Font.color Palette.primary5 ] [ text "admin" ]
                            ]
                        ]
                    , row
                        [ Border.widthEach { top = 1, left = 0, right = 0, bottom = 0 }
                        , Border.color Palette.neutral6
                        , mouseOver [ Background.color Palette.neutral4 ]
                        , width fill
                        , paddingXY 10 20
                        ]
                        [ text "Sign out" ]
                    ]

            else
                none
    in
    case maybeViewer of
        Just viewer ->
            [ el
                [ width (px 30)
                , height (px 30)
                , Border.rounded 180
                , Background.image (Asset.src Asset.defaultAvatar)
                , Font.size 16
                , pointer
                , below dropdownMenu
                , onClick (headerMsg (Header { userMenuOpen = not userMenuOpen }))
                , Border.shadow
                    { offset = ( 0, 0 )
                    , size =
                        if userMenuOpen then
                            5

                        else
                            0
                    , blur = 10
                    , color = Palette.neutral4
                    }
                ]
                none
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
            , color = Palette.neutral6
            }
        , Background.color Palette.neutral7
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
