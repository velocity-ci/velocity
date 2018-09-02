module Main exposing (main)

import Context exposing (Context)
import Data.Session as Session exposing (Session)
import Data.User as User exposing (User, Username)
import Navigation exposing (Location)
import Views.Page as Page exposing (ActivePage)
import Bootstrap.Dropdown as Dropdown
import Page.Errored as Errored exposing (PageLoadError)
import Page.Home as Home
import Page.Login as Login
import Page.NotFound as NotFound
import Page.Project as Project
import Page.KnownHosts as KnownHosts
import Page.Users as Users
import Page.Users as User
import Request.Errors
import Request.Channel
import Route exposing (Route)
import Util exposing ((=>))
import Phoenix.Socket as Socket exposing (Socket)
import Json.Decode as Decode exposing (Value)
import Task
import Ports
import Component.UserMenuDropdown as UserMenuDropdown
import Html exposing (Html, text, div)
import Component.Sidebar as Sidebar
import Window
import Animation


type Page
    = Blank
    | NotFound
    | Errored PageLoadError
    | Home Home.Model
    | Project Project.Model
    | Login Login.Model
    | KnownHosts KnownHosts.Model
    | Users Users.Model


type PageState
    = Loaded Page
    | TransitioningFrom Page



-- MODEL --


type alias Model =
    { session : Session Msg
    , context : Context
    , pageState : PageState
    , userDropdown : UserMenuDropdown.State
    , sidebarDisplayType : Sidebar.DisplayType
    , pageWidth : Int
    }


type alias ProgramFlags =
    { apiUrlBase : String
    , session : Value
    }


init : ProgramFlags -> Location -> ( Model, Cmd Msg )
init flags location =
    let
        user =
            decodeUserFromJson flags.session

        context =
            Context.initContext flags.apiUrlBase

        session =
            { user = user
            , socket = initialSocket context
            }

        maybeRoute =
            (Route.fromLocation location)

        ( initialModel, initialCmd ) =
            setRoute maybeRoute
                { pageState = Loaded initialPage
                , context = context
                , session = session
                , userDropdown = UserMenuDropdown.init
                , sidebarDisplayType = Sidebar.initDisplayType 0 Sidebar.normalSize
                , pageWidth = 0
                }
    in
        initialModel
            ! [ initialCmd
              , Task.perform WindowWidthChange Window.width
              ]


decodeUserFromJson : Value -> Maybe User
decodeUserFromJson json =
    json
        |> Decode.decodeValue Decode.string
        |> Result.toMaybe
        |> Maybe.andThen (Decode.decodeString User.decoder >> Result.toMaybe)


initialPage : Page
initialPage =
    Blank


initialSocket : Context -> Socket Msg
initialSocket { wsUrl } =
    Socket.init wsUrl
        |> Socket.withoutHeartbeat


userDropdownConfig : UserMenuDropdown.Config Msg
userDropdownConfig =
    { userDropdownMsg = UserDropdownToggleMsg
    , newUrlMsg = NewUrl
    }



-- VIEW --


view : Model -> Html Msg
view model =
    let
        page =
            viewPage model.sidebarDisplayType model.session

        sidebar =
            viewSidebar model
    in
        case model.pageState of
            Loaded activePage ->
                div []
                    [ sidebar False activePage
                    , page False activePage
                    ]

            TransitioningFrom activePage ->
                div []
                    [ sidebar True activePage
                    , page True activePage
                    ]


pageToActivePage : Page -> ActivePage
pageToActivePage page =
    case page of
        Home _ ->
            Page.Home

        Project _ ->
            Page.Project

        KnownHosts _ ->
            Page.KnownHosts

        _ ->
            Page.Other


viewSidebar : Model -> Bool -> Page -> Html Msg
viewSidebar model isLoading page =
    let
        pageSidebar =
            case page of
                Project subModel ->
                    subModel
                        |> Project.viewProjectNavigation
                        |> Html.map ProjectMsg

                _ ->
                    text ""

        userDropdown =
            case model.session.user of
                Just user ->
                    UserMenuDropdown.view model.userDropdown userDropdownConfig

                Nothing ->
                    text ""

        content =
            Page.sidebar [ pageSidebar, userDropdown ]

        subSidebar =
            case page of
                Project subModel ->
                    subModel
                        |> Project.viewSubpageProjectNavigation
                        |> Html.map ProjectMsg

                _ ->
                    text ""
    in
        Page.sidebarFrame model.sidebarDisplayType sidebarConfig content subSidebar


viewPage : Sidebar.DisplayType -> Session Msg -> Bool -> Page -> Html Msg
viewPage sidebarDisplayType session isLoading page =
    let
        frame =
            Page.frame isLoading session.user sidebarConfig sidebarDisplayType
    in
        case page of
            NotFound ->
                NotFound.view session
                    |> frame Page.Other

            Blank ->
                -- This is for the very initial page load, while we are loading
                -- data via HTTP. We could also render a spinner here.
                Html.text ""
                    |> frame Page.Other

            Errored subModel ->
                Errored.view session subModel
                    |> frame Page.Other

            Home subModel ->
                Home.view session subModel
                    |> Html.map HomeMsg
                    |> frame Page.Home

            Project subModel ->
                Project.view session subModel
                    |> Html.map ProjectMsg
                    |> frame Page.Projects

            Login subModel ->
                Login.view session subModel
                    |> Html.map LoginMsg
                    |> frame Page.Login

            KnownHosts subModel ->
                KnownHosts.view session subModel
                    |> Html.map KnownHostsMsg
                    |> frame Page.KnownHosts

            Users subModel ->
                Users.view session subModel
                    |> Html.map UsersMsg
                    |> frame Page.Users



-- SUBSCRIPTIONS --


sidebarConfig : Sidebar.Config Msg
sidebarConfig =
    { hideCollapsableSidebarMsg = HideSidebar
    , showCollapsableSidebarMsg = ShowSidebar
    , animateMsg = AnimateSidebar
    , newUrlMsg = NewUrl
    }


subscriptions : Model -> Sub Msg
subscriptions model =
    let
        session =
            Sub.map SetUser sessionChange

        socket =
            Socket.listen model.session.socket SocketMsg

        userDropdown =
            UserMenuDropdown.subscriptions userDropdownConfig model.userDropdown

        resizes =
            Window.resizes (.width >> WindowWidthChange)

        sidebar =
            Sidebar.subscriptions sidebarConfig model.sidebarDisplayType

        page =
            model.pageState
                |> getPage
                |> pageSubscriptions
    in
        Sub.batch
            [ session
            , socket
            , page
            , userDropdown
            , resizes
            , sidebar
            ]


pageSubscriptions : Page -> Sub Msg
pageSubscriptions page =
    case page of
        Project subModel ->
            Project.subscriptions subModel
                |> Sub.map ProjectMsg

        Home subModel ->
            Home.subscriptions subModel
                |> Sub.map HomeMsg

        KnownHosts subModel ->
            KnownHosts.subscriptions subModel
                |> Sub.map KnownHostsMsg

        Users subModel ->
            Users.subscriptions subModel
                |> Sub.map UsersMsg

        _ ->
            Sub.none


sessionChange : Sub (Maybe User)
sessionChange =
    Ports.onSessionChange (Decode.decodeValue User.decoder >> Result.toMaybe)


getPage : PageState -> Page
getPage pageState =
    case pageState of
        Loaded page ->
            page

        TransitioningFrom page ->
            page



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe Route)
    | HomeMsg Home.Msg
    | HomeLoaded Home.Model
    | SetUser (Maybe User)
    | SessionExpired
    | LoadFailed PageLoadError
    | LoginMsg Login.Msg
    | ProjectLoaded ( Project.Model, Cmd Project.Msg )
    | ProjectMsg Project.Msg
    | KnownHostsLoaded KnownHosts.Model
    | KnownHostsMsg KnownHosts.Msg
    | UsersLoaded Users.Model
    | UsersMsg Users.Msg
    | SocketMsg (Socket.Msg Msg)
    | UserDropdownToggleMsg Dropdown.State
    | WindowWidthChange Int
    | AnimateSidebar Animation.Msg
    | ShowSidebar
    | HideSidebar
    | NoOp


leavePageChannels : Session Msg -> Page -> Maybe Route -> ( Session Msg, Cmd Msg )
leavePageChannels session page route =
    let
        channels =
            channelsToLeaveOnRouteChange page route

        ( newSocket, leaveCmd ) =
            Request.Channel.leaveChannels SocketMsg channels session.socket
    in
        { session | socket = newSocket }
            => leaveCmd


channelsToLeaveOnRouteChange : Page -> (Maybe Route -> List String)
channelsToLeaveOnRouteChange page =
    case page of
        Home _ ->
            Home.leaveChannels

        Project subModel ->
            Project.leaveChannels subModel

        _ ->
            always []


handledErrorToMsg : Request.Errors.HandledError -> Msg
handledErrorToMsg err =
    case err of
        Request.Errors.Unauthorized ->
            SessionExpired


handledChannelErrorToMsg : Request.Errors.Error unhandled -> Msg
handledChannelErrorToMsg err =
    case err of
        Request.Errors.HandledError err ->
            handledErrorToMsg err

        _ ->
            NoOp



--setSidebar : Maybe Route -> Int -> Sidebar.DisplayType -> Sidebar.DisplayType
--setSidebar maybeRoute pageWidth displayType =
--    case maybeRoute of
--        Just (Route.Project _ projectRoute) ->
--            Project.setSidebar (Just projectRoute) pageWidth displayType
--
--        _ ->
--            Sidebar.initDisplayType pageWidth Sidebar.normalSize
--sidebarSize : Model -> Sidebar.Size
--sidebarSize model =
--    case getPage model.pageState of
--        Project subModel ->
--            Project.sidebarSize subModel
--
--        _ ->
--            Sidebar.normalSize


setRoute : Maybe Route -> Model -> ( Model, Cmd Msg )
setRoute maybeRoute model =
    let
        transition successMsg task =
            let
                model_ =
                    { model | pageState = TransitioningFrom (getPage model.pageState) }

                handleResult result =
                    case result of
                        Err (Request.Errors.HandledError handledError) ->
                            handledErrorToMsg handledError

                        Err (Request.Errors.UnhandledError pageLoadError) ->
                            LoadFailed pageLoadError

                        Ok payload ->
                            successMsg payload
            in
                model_ => Task.attempt handleResult task

        errored =
            pageErrored model

        session =
            model.session

        socket =
            session.socket

        maybeToken =
            Maybe.map .token session.user

        joinChannels =
            Request.Channel.joinChannels socket maybeToken handledChannelErrorToMsg
    in
        case maybeRoute of
            Nothing ->
                { model | pageState = Loaded NotFound } => Cmd.none

            Just (Route.Home) ->
                let
                    ( newModel, pageCmd ) =
                        transition HomeLoaded (Home.init model.context model.session)

                    ( listeningSocket, socketCmd ) =
                        joinChannels HomeMsg Home.initialEvents
                in
                    case model.session.user of
                        Just user ->
                            { newModel | session = { session | socket = listeningSocket } }
                                ! [ pageCmd, Cmd.map SocketMsg socketCmd ]

                        Nothing ->
                            model => Route.modifyUrl Route.Login

            Just (Route.Login) ->
                { model | pageState = Loaded (Login Login.initialModel) } => Cmd.none

            Just (Route.Users) ->
                case model.session.user of
                    Just user ->
                        transition UsersLoaded (Users.init model.context model.session)

                    Nothing ->
                        model => Route.modifyUrl Route.Login

            Just (Route.Logout) ->
                let
                    session =
                        model.session
                in
                    { model | session = { session | user = Nothing } }
                        ! [ Ports.storeSession Nothing
                          , Route.modifyUrl Route.Home
                          ]

            Just (Route.Projects) ->
                setRoute (Just Route.Home) model

            Just (Route.KnownHosts) ->
                case model.session.user of
                    Just user ->
                        transition KnownHostsLoaded (KnownHosts.init model.context model.session)

                    Nothing ->
                        model => Route.modifyUrl Route.Login

            Just (Route.Project slug subRoute) ->
                let
                    ( listeningSocket, socketCmd ) =
                        Project.initialEvents slug subRoute
                            |> joinChannels ProjectMsg

                    transitionSubPage subModel =
                        let
                            ( ( newModel, newMsg ), externalMsg ) =
                                Project.update model.context model.session (Project.SetRoute (Just subRoute)) subModel

                            externalUpdatedModel =
                                case externalMsg of
                                    Project.SetSidebarSize size ->
                                        { model | sidebarDisplayType = Sidebar.initDisplayType model.pageWidth size }

                                    Project.NoOp_ ->
                                        model
                        in
                            { externalUpdatedModel
                                | pageState = Loaded (Project newModel)
                                , session = { session | socket = listeningSocket }
                            }
                                ! [ Cmd.map ProjectMsg newMsg ]

                    ( pageModel, pageCmd ) =
                        Just subRoute
                            |> Project.init model.context model.session slug
                            |> transition ProjectLoaded
                            |> Tuple.mapFirst (\m -> { m | session = { session | socket = listeningSocket } })
                            |> Tuple.mapSecond (\c -> Cmd.batch [ c, Cmd.map SocketMsg socketCmd ])
                in
                    case ( model.session.user, model.pageState ) of
                        ( Just _, Loaded page ) ->
                            case page of
                                -- If we're on the product page for the same product as the new route just load sub-page
                                -- Otherwise load the project page fresh
                                Project subModel ->
                                    if slug == subModel.project.slug then
                                        transitionSubPage subModel
                                    else
                                        ( pageModel, pageCmd )

                                _ ->
                                    ( pageModel, pageCmd )

                        ( Just _, TransitioningFrom _ ) ->
                            ( pageModel, pageCmd )

                        ( Nothing, _ ) ->
                            model => Route.modifyUrl Route.Login


pageErrored : Model -> ActivePage -> String -> ( Model, Cmd msg )
pageErrored model activePage errorMessage =
    let
        error =
            Errored.pageLoadError activePage errorMessage
    in
        { model | pageState = Loaded (Errored error) } => Cmd.none


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    let
        ( updatedPageModel, pageCmd ) =
            updatePage (getPage model.pageState) msg model

        --
        --        updatedSidebar =
        --            updateSidebar (getPage updatedPageModel.pageState) model.pageWidth model.sidebarDisplayType
    in
        updatedPageModel
            --        { updatedPageModel | sidebarDisplayType = updatedSidebar }
            =>
                pageCmd


setRouteUpdate : Maybe Route -> Model -> ( Model, Cmd Msg )
setRouteUpdate maybeRoute model =
    let
        ( channelLeaveSession, channelLeaveCmd ) =
            leavePageChannels model.session (getPage model.pageState) maybeRoute

        ( routeModel, routeCmd ) =
            setRoute maybeRoute { model | session = channelLeaveSession }
    in
        routeModel
            ! [ routeCmd, channelLeaveCmd ]


sidebarSize : Page -> Sidebar.Size
sidebarSize page =
    case page of
        Project subModel ->
            Project.sidebarSize subModel

        _ ->
            Sidebar.normalSize


updateSidebar : Page -> Int -> Sidebar.DisplayType
updateSidebar page windowWidth =
    page
        |> sidebarSize
        |> Sidebar.initDisplayType windowWidth


updatePage : Page -> Msg -> Model -> ( Model, Cmd Msg )
updatePage page msg model =
    let
        session =
            model.session

        --        toPage toModel toMsg subUpdate subMsg subModel =
        --            let
        --                ( newModel, newCmd ) =
        --                    subUpdate subMsg subModel
        --
        --                page =
        --                    toModel newModel
        --            in
        --                { model
        --                    | pageState = Loaded page
        --                    , sidebarDisplayType = updateSidebar page model.pageWidth
        --                }
        --                    ! [ Cmd.map toMsg newCmd ]
        errored =
            pageErrored model

        maybeToken =
            Maybe.map .token session.user

        joinChannels =
            Request.Channel.joinChannels session.socket maybeToken handledChannelErrorToMsg
    in
        case ( msg, page ) of
            ( NewUrl url, _ ) ->
                model
                    => Navigation.newUrl url

            ( WindowWidthChange width, _ ) ->
                { model
                    | sidebarDisplayType = updateSidebar page model.pageWidth
                    , pageWidth = width
                }
                    => Cmd.none

            ( AnimateSidebar animateMsg, _ ) ->
                { model | sidebarDisplayType = Sidebar.animate model.sidebarDisplayType animateMsg }
                    => Cmd.none

            ( ShowSidebar, _ ) ->
                { model | sidebarDisplayType = Debug.log "WHAT" (Sidebar.show model.sidebarDisplayType) }
                    => Cmd.none

            ( HideSidebar, _ ) ->
                { model | sidebarDisplayType = Sidebar.hide model.sidebarDisplayType }
                    => Cmd.none

            ( SocketMsg msg, _ ) ->
                let
                    ( newSocket, socketCmd ) =
                        Socket.update msg model.session.socket
                in
                    ( { model | session = { session | socket = newSocket } }
                    , Cmd.map SocketMsg socketCmd
                    )

            ( SetRoute route, _ ) ->
                setRouteUpdate route model

            ( SessionExpired, _ ) ->
                setRouteUpdate (Just Route.Logout) model

            ( UserDropdownToggleMsg state, _ ) ->
                let
                    sidebar =
                        model.userDropdown
                in
                    { model | userDropdown = { sidebar | userDropdown = state } }
                        => Cmd.none

            ( SetUser user, _ ) ->
                let
                    session =
                        model.session

                    cmd =
                        -- If we just signed out, then redirect to Home.
                        if session.user /= Nothing && user == Nothing then
                            Route.modifyUrl Route.Home
                        else
                            Cmd.none
                in
                    { model
                        | session =
                            { session
                                | user = user
                            }
                    }
                        => cmd

            ( LoadFailed error, _ ) ->
                { model
                    | pageState = Loaded (Errored error)
                    , sidebarDisplayType = updateSidebar (Errored error) model.pageWidth
                }
                    => Cmd.none

            ( LoginMsg subMsg, Login subModel ) ->
                let
                    ( ( pageModel, cmd ), msgFromPage ) =
                        Login.update model.context subMsg subModel

                    newModel =
                        case msgFromPage of
                            Login.NoOp ->
                                model

                            Login.SetUser user ->
                                let
                                    session =
                                        model.session
                                in
                                    { model
                                        | session =
                                            { session
                                                | user = Just user
                                            }
                                    }
                in
                    { newModel | pageState = Loaded (Login pageModel) }
                        => Cmd.map LoginMsg cmd

            ( HomeLoaded subModel, _ ) ->
                { model
                    | pageState = Loaded (Home subModel)
                    , sidebarDisplayType = updateSidebar (Home subModel) model.pageWidth
                }
                    => Cmd.none

            ( HomeMsg subMsg, Home subModel ) ->
                let
                    ( ( newSubModel, newSubCmd ), externalMsg ) =
                        Home.update model.context session subMsg subModel

                    model_ =
                        { model | pageState = Loaded (Home newSubModel) }

                    ( modelAfterExternalMsg, cmdAfterExternalMsg ) =
                        case externalMsg of
                            Home.NoOp ->
                                model_ => Cmd.none

                            Home.HandleRequestError err ->
                                update (handledErrorToMsg err) model_
                in
                    modelAfterExternalMsg ! [ Cmd.map HomeMsg newSubCmd, cmdAfterExternalMsg ]

            ( UsersLoaded subModel, _ ) ->
                { model
                    | pageState = Loaded (Users subModel)
                    , sidebarDisplayType = updateSidebar (Users subModel) model.pageWidth
                }
                    => Cmd.none

            ( UsersMsg subMsg, Users subModel ) ->
                let
                    ( ( newSubModel, newSubCmd ), externalMsg ) =
                        Users.update model.context session subMsg subModel

                    model_ =
                        { model | pageState = Loaded (Users newSubModel) }

                    ( modelAfterExternalMsg, cmdAfterExternalMsg ) =
                        case externalMsg of
                            Users.NoOp ->
                                model_ => Cmd.none

                            Users.HandleRequestError err ->
                                update (handledErrorToMsg err) model_
                in
                    modelAfterExternalMsg
                        ! [ Cmd.map UsersMsg newSubCmd
                          , cmdAfterExternalMsg
                          ]

            ( KnownHostsLoaded subModel, _ ) ->
                { model
                    | pageState = Loaded (KnownHosts subModel)
                    , sidebarDisplayType = updateSidebar (KnownHosts subModel) model.pageWidth
                }
                    => Cmd.none

            ( KnownHostsMsg subMsg, KnownHosts subModel ) ->
                let
                    ( ( newSubModel, newSubCmd ), externalMsg ) =
                        KnownHosts.update model.context session subMsg subModel

                    model_ =
                        { model | pageState = Loaded (KnownHosts newSubModel) }

                    ( modelAfterExternalMsg, cmdAfterExternalMsg ) =
                        case externalMsg of
                            KnownHosts.NoOp ->
                                model_ => Cmd.none

                            KnownHosts.HandleRequestError err ->
                                update (handledErrorToMsg err) model_
                in
                    modelAfterExternalMsg ! [ Cmd.map KnownHostsMsg newSubCmd, cmdAfterExternalMsg ]

            ( ProjectLoaded ( subModel, subMsg ), _ ) ->
                let
                    pageState =
                        Loaded (Project subModel)
                in
                    { model
                        | pageState = pageState
                        , sidebarDisplayType = updateSidebar (Project subModel) model.pageWidth
                    }
                        ! [ Cmd.map ProjectMsg subMsg ]

            ( ProjectMsg subMsg, Project subModel ) ->
                let
                    session =
                        model.session

                    socket =
                        session.socket

                    ( ( newSubModel, newCmd ), _ ) =
                        Project.update model.context session subMsg subModel

                    ( listeningSocket, socketCmd ) =
                        Project.loadedEvents subMsg subModel
                            |> joinChannels ProjectMsg
                in
                    { model
                        | pageState = Loaded (Project newSubModel)
                        , session = { session | socket = listeningSocket }
                    }
                        ! [ Cmd.map ProjectMsg newCmd
                          , Cmd.map SocketMsg socketCmd
                          ]

            ( _, NotFound ) ->
                -- Disregard incoming messages when we're on the
                -- NotFound page.
                model => Cmd.none

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong page
                model => Cmd.none



-- MAIN --


main : Program ProgramFlags Model Msg
main =
    Navigation.programWithFlags (Route.fromLocation >> SetRoute)
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }
