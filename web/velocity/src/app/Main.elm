module Main exposing (main)

import Context exposing (Context)
import Data.Session as Session exposing (Session)
import Data.User as User exposing (User, Username)
import Navigation exposing (Location)
import Views.Page as Page exposing (ActivePage)
import Page.Errored as Errored exposing (PageLoadError)
import Page.Home as Home
import Page.Login as Login
import Page.NotFound as NotFound
import Page.Projects as Projects
import Page.Project as Project
import Page.KnownHosts as KnownHosts
import Request.Errors
import Request.Channel
import Route exposing (Route)
import Util exposing ((=>))
import Page.Header as Header
import Phoenix.Socket as Socket exposing (Socket)
import Html exposing (..)
import Json.Decode as Decode exposing (Value)
import Task
import Ports


type Page
    = Blank
    | NotFound
    | Errored PageLoadError
    | Home Home.Model
    | Projects Projects.Model
    | Project Project.Model
    | Login Login.Model
    | KnownHosts KnownHosts.Model


type PageState
    = Loaded Page
    | TransitioningFrom Page



-- MODEL --


type alias Model =
    { session : Session Msg
    , context : Context
    , pageState : PageState
    , headerState : Header.Model
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

        ( headerState, headerCmd ) =
            Header.init

        ( initialModel, initialCmd ) =
            setRoute (Route.fromLocation location)
                { pageState = Loaded initialPage
                , context = context
                , session = session
                , headerState = headerState
                }
    in
        initialModel
            ! [ Cmd.map HeaderMsg headerCmd
              , initialCmd
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



-- VIEW --


view : Model -> Html Msg
view model =
    let
        page =
            viewPage model.session

        header =
            viewHeader model
    in
        case model.pageState of
            Loaded activePage ->
                div []
                    [ header False activePage
                    , page False activePage
                    ]

            TransitioningFrom activePage ->
                div []
                    [ header True activePage
                    , page True activePage
                    ]


pageToActivePage : Page -> ActivePage
pageToActivePage page =
    case page of
        Home _ ->
            Page.Home

        Projects _ ->
            Page.Projects

        Project _ ->
            Page.Project

        KnownHosts _ ->
            Page.KnownHosts

        _ ->
            Page.Other


viewHeader : Model -> Bool -> Page -> Html Msg
viewHeader { session, headerState } isLoading page =
    page
        |> pageToActivePage
        |> Header.view headerState session.user isLoading
        |> Html.map HeaderMsg


viewPage : Session Msg -> Bool -> Page -> Html Msg
viewPage session isLoading page =
    let
        frame =
            Page.frame isLoading session.user
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

            Projects subModel ->
                Projects.view session subModel
                    |> Html.map ProjectsMsg
                    |> frame Page.Projects

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



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions model =
    let
        session =
            Sub.map SetUser sessionChange

        header =
            model.headerState
                |> Header.subscriptions
                |> Sub.map HeaderMsg

        socket =
            Socket.listen model.session.socket SocketMsg

        page =
            model.pageState
                |> getPage
                |> pageSubscriptions
    in
        Sub.batch [ header, session, socket, page ]


pageSubscriptions : Page -> Sub Msg
pageSubscriptions page =
    case page of
        Project subModel ->
            Project.subscriptions subModel
                |> Sub.map ProjectMsg

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
    = SetRoute (Maybe Route)
    | HomeMsg Home.Msg
    | HomeLoaded Home.Model
    | SetUser (Maybe User)
    | SessionExpired
    | LoadFailed PageLoadError
    | LoginMsg Login.Msg
    | ProjectsLoaded Projects.Model
    | ProjectsMsg Projects.Msg
    | ProjectLoaded ( Project.Model, Cmd Project.Msg )
    | ProjectMsg Project.Msg
    | KnownHostsLoaded KnownHosts.Model
    | KnownHostsMsg KnownHosts.Msg
    | SocketMsg (Socket.Msg Msg)
    | HeaderMsg Header.Msg
    | NoOp


leavePageChannels : Session Msg -> Page -> Maybe Route -> ( Session Msg, Cmd Msg )
leavePageChannels session page route =
    let
        leaveChannels =
            Request.Channel.leaveChannels SocketMsg

        ( newSocket, leaveCmd ) =
            case page of
                Projects _ ->
                    if route == Just Route.Projects then
                        session.socket => Cmd.none
                    else
                        leaveChannels [ Projects.channelName ] session.socket

                Home _ ->
                    if route == Just Route.Home then
                        session.socket => Cmd.none
                    else
                        leaveChannels [ Home.channelName ] session.socket

                Project subModel ->
                    case route of
                        Just (Route.Project projectSlug subRoute) ->
                            leaveChannels (Project.leaveChannels subModel (Just projectSlug) (Just subRoute)) session.socket

                        _ ->
                            leaveChannels (Project.leaveChannels subModel Nothing Nothing) session.socket

                _ ->
                    session.socket => Cmd.none
    in
        { session | socket = newSocket }
            => leaveCmd


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
                case model.session.user of
                    Just user ->
                        let
                            ( newModel, pageCmd ) =
                                transition ProjectsLoaded (Projects.init model.context model.session)

                            ( listeningSocket, socketCmd ) =
                                joinChannels ProjectsMsg Projects.initialEvents
                        in
                            { newModel | session = { session | socket = listeningSocket } }
                                ! [ pageCmd, Cmd.map SocketMsg socketCmd ]

                    Nothing ->
                        model => Route.modifyUrl Route.Login

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

        setRouteUpdate route model =
            let
                ( channelLeaveSession, channelLeaveCmd ) =
                    leavePageChannels model.session (getPage model.pageState) route

                ( routeModel, routeCmd ) =
                    setRoute route { model | session = channelLeaveSession }
            in
                routeModel
                    ! [ routeCmd, channelLeaveCmd ]

        errored =
            pageErrored model

        maybeToken =
            Maybe.map .token session.user

        joinChannels =
            Request.Channel.joinChannels session.socket maybeToken handledChannelErrorToMsg
    in
        case ( msg, page ) of
            ( SocketMsg msg, _ ) ->
                let
                    ( newSocket, socketCmd ) =
                        Socket.update msg model.session.socket
                in
                    ( { model | session = { session | socket = newSocket } }
                    , Cmd.map SocketMsg socketCmd
                    )

            ( HeaderMsg subMsg, _ ) ->
                let
                    ( headerState, headerCmd ) =
                        Header.update subMsg model.headerState
                in
                    { model | headerState = headerState }
                        => Cmd.map HeaderMsg headerCmd

            ( SetRoute route, _ ) ->
                setRouteUpdate route model

            ( SessionExpired, _ ) ->
                setRouteUpdate (Just Route.Logout) model

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
                toPage Home HomeMsg (Home.update session) subMsg subModel

            ( ProjectsLoaded subModel, _ ) ->
                { model | pageState = Loaded (Projects subModel) } => Cmd.none

            ( ProjectsMsg subMsg, Projects subModel ) ->
                let
                    ( ( newSubModel, newSubCmd ), externalMsg ) =
                        Projects.update model.context session subMsg subModel

                    model_ =
                        { model | pageState = Loaded (Projects newSubModel) }

                    ( modelAfterExternalMsg, cmdAfterExternalMsg ) =
                        case externalMsg of
                            Projects.NoOp ->
                                model_ => Cmd.none

                            Projects.HandleRequestError err ->
                                update (handledErrorToMsg err) model_
                in
                    modelAfterExternalMsg ! [ Cmd.map ProjectsMsg newSubCmd, cmdAfterExternalMsg ]

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
