module Page exposing (DropdownStatus(..), Layout, Page(..), initLayout, layoutSubscriptions, view, viewErrors)

import Activity
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
import Layout.Header as Header
import Palette
import Route exposing (Route)
import Username exposing (Username)
import Viewer exposing (Viewer)



-- Header


type DropdownStatus
    = Open
    | ListenClicks
    | Closed


type Layout
    = Layout DropdownStatus Bool


initLayout : Layout
initLayout =
    Layout Closed False


{-| The dropdowns makes use of subscriptions to ensure that opened dropdowns are
automatically closed when you click outside them.
-}
layoutSubscriptions : Layout -> (Layout -> msg) -> Sub msg
layoutSubscriptions (Layout status notificationsOpen) updateStatus =
    case status of
        Open ->
            Browser.Events.onAnimationFrame
                (\_ -> updateStatus (Layout ListenClicks notificationsOpen))

        ListenClicks ->
            Browser.Events.onClick
                (Decode.succeed (updateStatus (Layout Closed notificationsOpen)))

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
type alias Config msg =
    { viewer : Maybe Viewer
    , page : Page
    , title : String
    , content : Element msg
    , layout : Layout
    , updateLayout : Layout -> msg
    , context : Context msg
    , log : Maybe Activity.Log
    }


view : Config msg -> { title : String, body : List (Html msg) }
view config =
    { title = config.title ++ " - Conduit"
    , body =
        [ Element.layout
            [ Font.family
                [ Font.typeface "Roboto"
                , Font.sansSerif
                ]
            , inFront (viewHeader config)
            , inFront (viewFooter config)
            ]
            (Element.row
                [ width fill
                , height fill
                , inFront
                    (viewIfDeviceIn config
                        [ Device Phone Portrait
                        , Device Phone Landscape
                        ]
                        (config.log
                            |> Maybe.map (Activity.view >> viewCollapsableNotificationsPanel config)
                            |> Maybe.withDefault none
                        )
                    )
                , inFront
                    (viewIfDeviceIn config
                        [ Device Tablet Portrait
                        , Device Tablet Landscape
                        ]
                        (column
                            [ width (fillPortion 1 |> maximum 300 |> minimum 250)
                            , height fill
                            , alignRight
                            ]
                            [ config.log
                                |> Maybe.map (Activity.view >> viewCollapsableNotificationsPanel config)
                                |> Maybe.withDefault none
                            ]
                        )
                    )
                ]
                [ el [ width (fillPortion 3), height fill ] (viewBody config.content)
                , viewIfDeviceIn config
                    [ Device Desktop Landscape
                    , Device Desktop Portrait
                    , Device BigDesktop Landscape
                    , Device BigDesktop Portrait
                    ]
                    (column
                        [ width (fillPortion 1 |> maximum 500 |> minimum 250)
                        , height fill
                        ]
                        [ config.log
                            |> Maybe.map Activity.view
                            |> Maybe.withDefault none
                        ]
                    )
                ]
            )
        ]
    }



-- Notifications


viewCollapsableNotificationsPanel : Config msg -> Element msg -> Element msg
viewCollapsableNotificationsPanel config content =
    let
        (Layout userMenu open) =
            config.layout
    in
    if open then
        content

    else
        none


viewIfDeviceIn : Config msg -> List Device -> Element msg -> Element msg
viewIfDeviceIn config devices content =
    if List.member (Context.device config.context) devices then
        content

    else
        none


viewIfDeviceNotIn : Config msg -> List Device -> Element msg -> Element msg
viewIfDeviceNotIn config devices content =
    if List.member (Context.device config.context) devices then
        none

    else
        content


viewBody : Element msg -> Element msg
viewBody content =
    row
        [ width fill
        , height fill
        , paddingEach { top = 60, bottom = 70, left = 0, right = 0 }
        , centerX
        ]
        [ content ]


viewHeader : Config msg -> Element msg
viewHeader config =
    if List.member (.class (Context.device config.context)) [ Phone, Tablet ] then
        viewMobileHeader config

    else
        viewDesktopHeader config


viewMobileHeader : Config msg -> Element msg
viewMobileHeader config =
    row
        [ width fill
        , height (px 60)
        , paddingXY 0 15
        , Background.color Palette.primary1

        --        , Background.color Palette.neutral7
        ]
        [ row
            [ width fill
            , centerX
            , height fill
            , inFront
                (column
                    [ alignRight
                    , moveLeft 10
                    , height fill
                    ]
                    [ column []
                        (viewMobileHeaderMenu config)
                    ]
                )
            ]
            [ el
                [ centerX ]
                viewBrand
            ]
        ]


viewDesktopHeader : Config msg -> Element msg
viewDesktopHeader { viewer, updateLayout, layout, page } =
    row
        [ width fill
        , height (px 60)
        , paddingXY 0 15
        , Background.color Palette.primary1
        ]
        [ row
            [ alignRight
            , width (fill |> maximum 2100)
            , paddingXY 20 0
            , height fill
            ]
            [ el [ alignLeft ] viewBrand
            , row
                [ alignRight
                , spacing 20
                , height fill
                ]
              <|
                viewDesktopHeaderMenu page viewer updateLayout layout
            ]
        ]


viewFooter : Config msg -> Element msg
viewFooter config =
    if List.member (.class (Context.device config.context)) [ Phone, Tablet ] then
        viewMobileFooter config

    else
        viewDesktopFooter config


viewDesktopFooter : Config msg -> Element msg
viewDesktopFooter config =
    el
        [ width fill
        , height (px 70)
        , paddingXY 20 15
        , alignBottom
        ]
        none


viewMobileFooter : Config msg -> Element msg
viewMobileFooter { context, page, viewer, updateLayout, layout } =
    let
        (Layout status notificationsPanel) =
            layout
    in
    column
        [ width fill
        , height (px 70)
        , paddingXY 20 15
        , alignBottom
        , Background.color Palette.transparent
        ]
        [ row
            [ centerY
            , centerX
            , width (fill |> maximum maxWidth)
            ]
            [ el
                [ width (px 35)
                , height (px 35)
                , Border.rounded 180
                , Background.image (Asset.src Asset.defaultAvatar)
                , Font.size 16
                , pointer
                , alignRight
                , above
                    (if status == ListenClicks then
                        Header.userMenuToggle

                     else
                        none
                    )
                , onClick
                    (if status == Closed then
                        updateLayout (Layout Open notificationsPanel)

                     else
                        updateLayout (Layout status notificationsPanel)
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
        ]


viewBrand : Element msg
viewBrand =
    el
        [ Font.color Palette.primary6
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


viewMobileHeaderMenu : Config msg -> List (Element msg)
viewMobileHeaderMenu config =
    let
        (Layout userMenu notificationsOpen) =
            config.layout

        linkTo =
            navbarLink config.page
    in
    case config.viewer of
        Just viewer ->
            [ Header.notificationsToggle
                { amount = 2
                , toggled = notificationsOpen
                , toggleMsg = Layout userMenu >> config.updateLayout
                }
            ]

        Nothing ->
            [ linkTo Route.Login (text "Sign in")
            ]


viewDesktopHeaderMenu : Page -> Maybe Viewer -> (Layout -> msg) -> Layout -> List (Element msg)
viewDesktopHeaderMenu page maybeViewer layoutMsg (Layout status notificationsPanel) =
    let
        linkTo =
            navbarLink page
    in
    case maybeViewer of
        Just viewer ->
            [ el
                [ width (px 30)
                , height (px 30)
                , alignTop
                , Border.rounded 180
                , Background.image (Asset.src Asset.defaultAvatar)
                , Font.size 16
                , pointer
                , below
                    (if status == ListenClicks then
                        Header.userMenuToggle

                     else
                        none
                    )
                , onClick
                    (if status == Closed then
                        layoutMsg (Layout Open notificationsPanel)

                     else
                        layoutMsg (Layout status notificationsPanel)
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
