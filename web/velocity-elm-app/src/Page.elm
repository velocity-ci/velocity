module Page exposing (Header, Page(..), headerSubscriptions, initHeader, view, viewErrors)

import Api exposing (Cred)
import Asset
import Browser exposing (Document)
import Browser.Events
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html)
import Icon
import Json.Decode as Decode
import Palette
import Route exposing (Route)
import Username exposing (Username)
import Viewer exposing (Viewer)



-- Header


type DropdownStatus
    = Open
    | ListenClicks
    | Closed


type Header
    = Header DropdownStatus


initHeader : Header
initHeader =
    Header Closed


{-| The dropdowns makes use of subscriptions to ensure that opened dropdowns are
automatically closed when you click outside them.
-}
headerSubscriptions : Header -> (Header -> msg) -> Sub msg
headerSubscriptions (Header status) updateStatus =
    case status of
        Open ->
            Browser.Events.onAnimationFrame
                (\_ -> updateStatus (Header ListenClicks))

        ListenClicks ->
            Browser.Events.onClick
                (Decode.succeed (updateStatus (Header Closed)))

        Closed ->
            Sub.none



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
    , context : Context
    }


view : Config subMsg msg -> { title : String, body : List (Html msg) }
view { context, viewer, page, title, content, toMsg, header, updateHeader } =
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
                [ viewHeader context page viewer updateHeader header
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


viewHeader : Context -> Page -> Maybe Viewer -> (Header -> msg) -> Header -> Element msg
viewHeader context page maybeViewer headerMsg header =
    let
        deviceClass =
            context
                |> Context.device
                |> .class
    in
    if deviceClass == Phone then
        --        viewMobileHeader page maybeViewer headerMsg header
        viewDesktopHeader page maybeViewer headerMsg header

    else
        viewDesktopHeader page maybeViewer headerMsg header


viewDesktopHeader : Page -> Maybe Viewer -> (Header -> msg) -> Header -> Element msg
viewDesktopHeader page maybeViewer headerMsg header =
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
            , height fill
            ]
            [ el [ alignLeft ] viewBrand
            , row
                [ centerY
                , alignRight
                , spacing 20
                , height fill
                ]
              <|
                viewMenu page maybeViewer headerMsg header
            ]
        ]


viewMobileHeader : Page -> Maybe Viewer -> (Header -> msg) -> Header -> Element msg
viewMobileHeader page maybeViewer headerMsg header =
    column
        [ width fill
        , height shrink
        , Border.shadow
            { offset = ( 0, 2 )
            , size = 2
            , blur = 2
            , color = Palette.neutral6
            }
        , Background.color Palette.neutral7
        , spacing 5
        ]
        [ row
            [ centerX
            , Border.widthEach { top = 0, left = 0, right = 0, bottom = 1 }
            , Border.color Palette.neutral6
            , width fill
            , paddingXY 0 10
            ]
            [ el [ centerX ] viewBrand
            ]
        , row
            [ width fill
            , padding 10
            ]
            [ column [ width fill ] []
            , column [ width fill ]
                [ paragraph [ Font.size 15, Font.color Palette.primary2 ] [ text "Signed in as" ]
                , paragraph [ Font.size 18, Font.heavy, Font.color Palette.primary5 ] [ text "admin" ]
                ]
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


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions


viewMenu : Page -> Maybe Viewer -> (Header -> msg) -> Header -> List (Element msg)
viewMenu page maybeViewer headerMsg (Header status) =
    let
        linkTo =
            navbarLink page

        dropdownMenu =
            case status of
                ListenClicks ->
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
                            , mouseOver
                                [ Background.color Palette.neutral2
                                , Font.color Palette.white
                                ]
                            , width fill
                            , paddingXY 20 20
                            , spacingXY 10 0
                            , Font.color Palette.primary1
                            , Font.light
                            , Font.size 16
                            ]
                            [ column
                                [ width shrink ]
                                [ Icon.logOut { iconOptions | size = 16 } ]
                            , column
                                [ width fill ]
                                [ text "Sign out" ]
                            ]
                        ]

                _ ->
                    none
    in
    case maybeViewer of
        Just viewer ->
            [ el
                [ height fill
                , Border.widthEach { top = 0, left = 0, right = 0, bottom = 3 }
                , Border.color Palette.transparent
                , paddingXY 8 0
                , pointer
                , Font.color Palette.neutral2
                , mouseOver
                    [ Border.color Palette.primary3
                    , Font.color Palette.primary3
                    ]
                ]
                (el
                    [ width (px 28)
                    , height (px 28)
                    , moveDown 3
                    , Border.rounded 180
                    , centerY
                    , above
                        (el
                            [ width shrink
                            , height shrink
                            , Background.color Palette.primary3
                            , Border.width 1
                            , Border.color Palette.primary7
                            , paddingXY 4 3
                            , Border.rounded 7
                            , moveDown 14
                            , moveRight 14
                            , Font.size 10
                            , Font.color Palette.white
                            ]
                            (text "2")
                        )
                    ]
                    (Icon.bell { iconOptions | size = 100, sizeUnit = Icon.Percentage })
                )
            , el
                [ width (px 30)
                , height (px 30)
                , Border.rounded 180
                , Background.image (Asset.src Asset.defaultAvatar)
                , Font.size 16
                , pointer
                , below dropdownMenu
                , onClick
                    (if status == Closed then
                        headerMsg (Header Open)

                     else
                        headerMsg (Header status)
                    )
                , Border.shadow
                    { offset = ( 0, 0 )
                    , size =
                        if status == ListenClicks then
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
