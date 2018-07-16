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
import Component.UserSidebar as UserSidebar
import Html exposing (Html, text, div)


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
    , userSidebar : UserSidebar.State
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

        ( initialModel, initialCmd ) =
            setRoute (Route.fromLocation location)
                { pageState = Loaded initialPage
                , context = context
                , session = session
                , userSidebar = UserSidebar.init
                }
    in
        initialModel
            ! [ initialCmd
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


userSidebarConfig : UserSidebar.Config Msg
userSidebarConfig =
    { userDropdownMsg = UserDropdownToggleMsg
    , newUrlMsg = NewUrl
    }



-- VIEW --


view : Model -> Html Msg
view model =
    let
        page =
            viewPage model.session

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
                    Project.viewSidebar model.session subModel
                        |> Html.map ProjectMsg

                _ ->
                    text ""

        userSidebar =
            case model.session.user of
                Just user ->
                    UserSidebar.view model.userSidebar userSidebarConfig

                Nothing ->
                    text ""
    in
        div [] [ pageSidebar, userSidebar ]
            |> Page.sidebarFrame NewUrl


viewPage : Session Msg -> Bool -> Page -> Html Msg
viewPage session isLoading page =
    let
        frame =
            Page.frame isLoading session.user
    in
        case page of
            NotFound ->
                NotFound.view session
                    |> frame Page.Other Page.NoSidebar

            Blank ->
                -- This is for the very initial page load, while we are loading
                -- data via HTTP. We could also render a spinner here.
                Html.text ""
                    |> frame Page.Other Page.NoSidebar

            Errored subModel ->
                Errored.view session subModel
                    |> frame Page.Other Page.NormalSidebar

            Home subModel ->
                Home.view session subModel
                    |> Html.map HomeMsg
                    |> frame Page.Home Page.NormalSidebar

            Project subModel ->
                let
                    sidebar =
                        if Project.hasExtraWideSidebar subModel session then
                            Page.ExtraWideSidebar
                        else
                            Page.NormalSidebar
                in
                    Project.view session subModel
                        |> Html.map ProjectMsg
                        |> frame Page.Projects sidebar

            Login subModel ->
                Login.view session subModel
                    |> Html.map LoginMsg
                    |> frame Page.Login Page.NormalSidebar

            KnownHosts subModel ->
                KnownHosts.view session subModel
                    |> Html.map KnownHostsMsg
                    |> frame Page.KnownHosts Page.NormalSidebar

            Users subModel ->
                Users.view session subModel
                    |> Html.map (always NoOp)
                    |> frame Page.Users Page.NormalSidebar



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions model =
    let
        session =
            Sub.map SetUser sessionChange

        socket =
            Socket.listen model.session.socket SocketMsg

        userSidebar =
            UserSidebar.subscriptions userSidebarConfig model.userSidebar

        page =
            model.pageState
                |> getPage
                |> pageSubscriptions
    in
        Sub.batch
            [ session
            , socket
            , page
            , userSidebar
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
                            ( newModel, newMsg ) =
                                Project.update model.context model.session (Project.SetRoute (Just subRoute)) subModel
                        in
                            { model
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
    updatePage (getPage model.pageState) msg model


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


updatePage : Page -> Msg -> Model -> ( Model, Cmd Msg )
updatePage page msg model =
    let
        session =
            model.session

        toPage toModel toMsg subUpdate subMsg subModel =
            let
                ( newModel, newCmd ) =
                    subUpdate subMsg subModel
            in
                { model | pageState = Loaded (toModel newModel) }
                    ! [ Cmd.map toMsg newCmd ]

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
                        model.userSidebar
                in
                    { model | userSidebar = { sidebar | userDropdown = state } }
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
                { model | pageState = Loaded (Errored error) } => Cmd.none

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
                { model | pageState = Loaded (Home subModel) } => Cmd.none

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
                { model | pageState = Loaded (Users subModel) }
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
                    modelAfterExternalMsg ! [ Cmd.map UsersMsg newSubCmd, cmdAfterExternalMsg ]

            ( KnownHostsLoaded subModel, _ ) ->
                { model | pageState = Loaded (KnownHosts subModel) } => Cmd.none

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
                    { model | pageState = pageState }
                        ! [ Cmd.map ProjectMsg subMsg
                          ]

            ( ProjectMsg subMsg, Project subModel ) ->
                let
                    session =
                        model.session

                    socket =
                        session.socket

                    ( newSubModel, newCmd ) =
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
