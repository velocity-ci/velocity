module Main exposing (main)

import Data.Session as Session exposing (Session)
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.User as User exposing (User, Username)
import Data.Project exposing (idToString)
import Html exposing (..)
import Json.Decode as Decode exposing (Value)
import Navigation exposing (Location)
import Page.Errored as Errored exposing (PageLoadError)
import Page.Home as Home
import Page.Login as Login
import Page.NotFound as NotFound
import Page.Projects as Projects
import Page.Project as Project
import Page.KnownHosts as KnownHosts
import Ports
import Route exposing (Route)
import Task
import Util exposing ((=>))
import Views.Page as Page exposing (ActivePage)
import Page.Header as Header
import Socket.Socket as Socket exposing (Socket)
import Socket.Channel as Channel exposing (Channel)


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
    , pageState : PageState
    }


init : Value -> Location -> ( Model, Cmd Msg )
init val location =
    let
        user =
            decodeUserFromJson val

        socket =
            Maybe.map (initialSocket << .token) user

        session =
            { user = user
            , socket = socket
            }
    in
        setRoute (Route.fromLocation location)
            { pageState = Loaded initialPage
            , session = session
            }


decodeUserFromJson : Value -> Maybe User
decodeUserFromJson json =
    json
        |> Decode.decodeValue Decode.string
        |> Result.toMaybe
        |> Maybe.andThen (Decode.decodeString User.decoder >> Result.toMaybe)


initialPage : Page
initialPage =
    Blank


initialSocket : AuthToken -> Socket Msg
initialSocket token =
    Socket.init <| "ws://localhost/v1/ws?authToken=" ++ AuthToken.tokenToString token



-- VIEW --


view : Model -> Html Msg
view model =
    let
        page =
            viewPage model.session

        header =
            viewHeader model.session
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

        _ ->
            Page.Other


viewHeader : Session msg -> Bool -> Page -> Html Msg
viewHeader session isLoading page =
    page
        |> pageToActivePage
        |> Header.viewHeader session.user isLoading
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

        --
        --        socket =
        --            Just SocketMsg
        --                |> Maybe.map2 Socket.listen model.socket
        --                |> Maybe.withDefault Sub.none
    in
        Sub.batch [ session ]


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
    = HeaderMsg Header.ExternalMsg
    | SetRoute (Maybe Route)
    | HomeMsg Home.Msg
    | HomeLoaded (Result PageLoadError Home.Model)
    | SetUser (Maybe User)
    | LoginMsg Login.Msg
    | ProjectsLoaded (Result PageLoadError Projects.Model)
    | ProjectsMsg Projects.Msg
    | ProjectLoaded (Result PageLoadError ( Project.Model, Cmd Project.Msg ))
    | ProjectMsg Project.Msg
    | KnownHostsLoaded (Result PageLoadError KnownHosts.Model)
    | KnownHostsMsg KnownHosts.Msg
    | SocketMsg (Socket.Msg Msg)
    | JoinChannel (Channel Msg)
    | NoOp



--    | NewMessage String


setRoute : Maybe Route -> Model -> ( Model, Cmd Msg )
setRoute maybeRoute model =
    let
        transition toMsg task =
            { model | pageState = TransitioningFrom (getPage model.pageState) }
                => Task.attempt toMsg task

        errored =
            pageErrored model
    in
        case maybeRoute of
            Nothing ->
                { model | pageState = Loaded NotFound } => Cmd.none

            Just (Route.Home) ->
                case model.session.user of
                    Just user ->
                        transition HomeLoaded (Home.init model.session)

                    Nothing ->
                        model => Route.modifyUrl Route.Login

            Just (Route.Login) ->
                { model | pageState = Loaded (Login Login.initialModel) } => Cmd.none

            Just (Route.Logout) ->
                let
                    session =
                        model.session
                in
                    { model | session = { session | user = Nothing, socket = Nothing } }
                        => Cmd.batch
                            [ Ports.storeSession Nothing
                            , Route.modifyUrl Route.Home
                            ]

            Just (Route.Projects) ->
                case model.session.user of
                    Just user ->
                        transition ProjectsLoaded (Projects.init model.session)

                    Nothing ->
                        errored Page.Projects "You must be signed in to access your projects."

            Just (Route.KnownHosts) ->
                case model.session.user of
                    Just user ->
                        transition KnownHostsLoaded (KnownHosts.init model.session)

                    Nothing ->
                        errored Page.KnownHosts "You must be signed in to access your known hosts."

            Just (Route.Project id subRoute) ->
                let
                    loadFreshPage =
                        Just subRoute
                            |> Project.init model.session id
                            |> Task.andThen
                                (\( ( model, cmd ), externalMsg ) ->
                                    let
                                        something =
                                            case externalMsg of
                                                Project.JoinChannel channel ->
                                                    Nothing

                                                _ ->
                                                    Nothing
                                    in
                                        Task.succeed ( model, cmd )
                                )
                            |> transition ProjectLoaded

                    transitionSubPage subModel =
                        let
                            ( ( newModel, newMsg ), externalMsg ) =
                                subModel
                                    |> Project.update model.session (Project.SetRoute (Just subRoute))
                        in
                            { model | pageState = Loaded (Project newModel) }
                                => Cmd.map ProjectMsg newMsg
                in
                    case ( model.session.user, model.pageState ) of
                        ( Just _, Loaded page ) ->
                            case page of
                                -- If we're on the product page for the same product as the new route just load sub-page
                                -- Otherwise load the project page fresh
                                Project subModel ->
                                    if id == subModel.project.id then
                                        transitionSubPage subModel
                                    else
                                        loadFreshPage

                                _ ->
                                    loadFreshPage

                        ( Just _, TransitioningFrom _ ) ->
                            loadFreshPage

                        ( Nothing, _ ) ->
                            errored Page.Project ("You must be signed in to access project '" ++ idToString id ++ "'.")


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

        errored =
            pageErrored model
    in
        case ( msg, page ) of
            ( SocketMsg msg, _ ) ->
                case model.session.socket of
                    Just socket ->
                        let
                            ( newSocket, socketCmd ) =
                                Socket.update msg socket
                        in
                            ( { model | session = { session | socket = Just newSocket } }
                            , Cmd.map SocketMsg socketCmd
                            )

                    Nothing ->
                        model => Cmd.none

            ( HeaderMsg subMsg, _ ) ->
                case subMsg of
                    Header.NewUrl newUrl ->
                        model => Navigation.newUrl newUrl

            ( SetRoute route, _ ) ->
                setRoute route model

            ( JoinChannel channel, _ ) ->
                let
                    session =
                        model.session

                    ( newSession, socketCmd ) =
                        case session.socket of
                            Just socket ->
                                let
                                    ( newSocket, socketCmd ) =
                                        Socket.join channel socket
                                in
                                    { session | socket = Just newSocket }
                                        => socketCmd

                            Nothing ->
                                session => Cmd.none
                in
                    { model | session = newSession }
                        => Cmd.map SocketMsg socketCmd

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
                                , socket = Maybe.map (initialSocket << .token) user
                            }
                    }
                        => cmd

            ( LoginMsg subMsg, Login subModel ) ->
                let
                    ( ( pageModel, cmd ), msgFromPage ) =
                        Login.update subMsg subModel

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
                                            { user = Just user
                                            , socket = Maybe.map (initialSocket << .token) (Just user)
                                            }
                                    }
                in
                    { newModel | pageState = Loaded (Login pageModel) }
                        => Cmd.map LoginMsg cmd

            ( HomeLoaded (Ok subModel), _ ) ->
                { model | pageState = Loaded (Home subModel) } => Cmd.none

            ( HomeLoaded (Err error), _ ) ->
                { model | pageState = Loaded (Errored error) } => Cmd.none

            ( HomeMsg subMsg, Home subModel ) ->
                toPage Home HomeMsg (Home.update session) subMsg subModel

            ( ProjectsLoaded (Ok subModel), _ ) ->
                { model | pageState = Loaded (Projects subModel) } => Cmd.none

            ( ProjectsLoaded (Err error), _ ) ->
                { model | pageState = Loaded (Errored error) } => Cmd.none

            ( ProjectsMsg subMsg, Projects subModel ) ->
                toPage Projects ProjectsMsg (Projects.update session) subMsg subModel

            ( KnownHostsLoaded (Ok subModel), _ ) ->
                { model | pageState = Loaded (KnownHosts subModel) } => Cmd.none

            ( KnownHostsLoaded (Err error), _ ) ->
                { model | pageState = Loaded (Errored error) } => Cmd.none

            ( KnownHostsMsg subMsg, KnownHosts subModel ) ->
                toPage KnownHosts KnownHostsMsg (KnownHosts.update session) subMsg subModel

            ( ProjectLoaded (Ok ( subModel, subMsg )), _ ) ->
                let
                    pageState =
                        Loaded (Project subModel)
                in
                    { model | pageState = pageState }
                        ! [ Cmd.map ProjectMsg subMsg
                          ]

            ( ProjectLoaded (Err error), _ ) ->
                { model | pageState = Loaded (Errored error) } => Cmd.none

            ( ProjectMsg subMsg, Project subModel ) ->
                let
                    ( ( newSubModel, newCmd ), externalMsg ) =
                        Project.update session subMsg subModel

                    ( newSession, socketCmd ) =
                        case externalMsg of
                            Project.NoOp ->
                                session
                                    => Cmd.none

                            Project.SetSocket ( socket, socketCmd_ ) ->
                                session
                                    => Cmd.none

                            Project.JoinChannel channel ->
                                case session.socket of
                                    Just socket ->
                                        let
                                            ( newSocket, socketCmd ) =
                                                Socket.join (Channel.map ProjectMsg channel) socket
                                        in
                                            { session | socket = Just newSocket }
                                                => socketCmd

                                    Nothing ->
                                        session => Cmd.none
                in
                    { model
                        | pageState = Loaded (Project newSubModel)
                        , session = newSession
                    }
                        ! [ Cmd.map ProjectMsg newCmd
                          ]

            ( _, NotFound ) ->
                -- Disregard incoming messages when we're on the
                -- NotFound page.
                model => Cmd.none

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong page
                model => Cmd.none



-- MAIN --


main : Program Value Model Msg
main =
    Navigation.programWithFlags (Route.fromLocation >> SetRoute)
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }
