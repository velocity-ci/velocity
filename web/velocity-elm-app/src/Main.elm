module Main exposing (Model(..), Msg(..), changeRouteTo, init, main, toSession, update, updateWith, view)

import Activity
import Api
import Api.Endpoint as Endpoint exposing (Endpoint)
import Browser exposing (Document)
import Browser.Events
import Browser.Navigation as Nav
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Json.Decode as Decode exposing (Decoder, Value, decodeString, field, string)
import Loading
import Page exposing (Layout)
import Page.Blank as BlankPage
import Page.Build as BuildPage
import Page.Home as HomePage
import Page.Home.ActivePanel as ActivePanel
import Page.Login as LoginPage
import Page.NotFound as NotFoundPage
import Page.Project as ProjectPage
import Phoenix.Socket as Socket
import Project
import Route exposing (Route)
import Session exposing (Session)
import Task exposing (Task)
import Url exposing (Url)
import Viewer exposing (Viewer)



---- MODEL ----


type Model
    = InitError String
    | InitialHTTPRequests (Context Msg) Nav.Key
    | ApplicationStarted Layout Body


type Body
    = Redirect Session (Context Msg)
    | NotFound Session (Context Msg)
    | Home (HomePage.Model Msg)
    | Login (LoginPage.Model Msg)
    | Project (ProjectPage.Model Msg)
    | Build (BuildPage.Model Msg)


init : Maybe Viewer -> Result Decode.Error (Context Msg) -> Url -> Nav.Key -> ( Model, Cmd Msg )
init maybeViewer contextResult url navKey =
    case contextResult of
        Ok context ->
            ( InitialHTTPRequests context navKey
            , Session.fromViewer navKey context maybeViewer
                |> Task.map (\session -> changeRouteTo (Route.fromUrl url) (Redirect session context))
                |> Task.attempt StartApplication
            )

        Err error ->
            ( InitError (Decode.errorToString error), Cmd.none )



-- VIEW


viewCurrentPage : Layout -> Body -> Document Msg
viewCurrentPage layout currentPage =
    let
        session =
            toSession currentPage

        viewPage page toMsg { title, content } =
            Page.view
                { viewer = Session.viewer session
                , page = page
                , title = title
                , content = Element.map toMsg content
                , layout = layout
                , updateLayout = UpdateLayout
                , context = toContext currentPage
                , log = activityLog session
                }
    in
    case currentPage of
        Redirect _ _ ->
            viewPage Page.Other (always Ignored) BlankPage.view

        NotFound _ _ ->
            viewPage Page.Other (always Ignored) NotFoundPage.view

        Home page ->
            viewPage Page.Home GotHomeMsg (HomePage.view page)

        Login page ->
            viewPage Page.Login GotLoginMsg (LoginPage.view page)

        Project page ->
            viewPage Page.Project GotProjectMsg (ProjectPage.view page)

        Build page ->
            viewPage Page.Build GotBuildMsg (BuildPage.view page)


activityLog : Session -> Activity.ViewConfiguration
activityLog session =
    Maybe.map2 Activity.ViewConfiguration (Session.log session) (Just <| Session.projects session)
        |> Maybe.withDefault { activities = Activity.init, projects = [] }


view : Model -> Document Msg
view model =
    case model of
        ApplicationStarted header currentPage ->
            viewCurrentPage header currentPage

        InitialHTTPRequests _ _ ->
            { title = "Loading"
            , body =
                [ layout
                    [ width fill, height fill ]
                    (el [ centerX, moveDown 100, width shrink, height shrink ] <|
                        Loading.icon { width = 50, height = 50 }
                    )
                ]
            }

        InitError error ->
            { title = "Context Error"
            , body =
                [ layout
                    [ Font.family
                        [ Font.typeface "Open Sans"
                        , Font.sansSerif
                        ]
                    , width fill
                    , height fill
                    ]
                    (textColumn [ alignLeft ] [ paragraph [] [ text error ] ])
                ]
            }



---- UPDATE ----


type Msg
    = Ignored
    | StartApplication (Result Session.InitError ( Body, Cmd Msg ))
    | ChangedRoute (Maybe Route)
    | ChangedUrl Url
    | ClickedLink Browser.UrlRequest
    | GotHomeMsg HomePage.Msg
    | GotLoginMsg LoginPage.Msg
    | GotProjectMsg ProjectPage.Msg
    | GotBuildMsg BuildPage.Msg
    | UpdateSession (Task Session.InitError Session)
    | UpdatedSession (Result Session.InitError Session)
    | WindowResized Int Int
    | UpdateLayout Page.Layout
    | SocketMsg (Socket.Msg Msg)
    | SocketUpdate Session.SocketUpdate
    | SetSession Session


toSession : Body -> Session
toSession page =
    case page of
        Redirect session _ ->
            session

        NotFound session _ ->
            session

        Home home ->
            HomePage.toSession home

        Login login ->
            LoginPage.toSession login

        Project project ->
            ProjectPage.toSession project

        Build build ->
            BuildPage.toSession build


toContext : Body -> Context Msg
toContext page =
    case page of
        Redirect _ context ->
            context

        NotFound _ context ->
            context

        Home home ->
            HomePage.toContext home

        Login login ->
            LoginPage.toContext login

        Project project ->
            ProjectPage.toContext project

        Build build ->
            BuildPage.toContext build


changeRouteTo : Maybe Route -> Body -> ( Body, Cmd Msg )
changeRouteTo maybeRoute currentPage =
    let
        session =
            toSession currentPage

        context =
            toContext currentPage
    in
    case maybeRoute of
        Nothing ->
            ( NotFound session context, Cmd.none )

        Just Route.Root ->
            ( currentPage
            , Route.replaceUrl (Session.navKey session) (Route.Home ActivePanel.None)
            )

        Just Route.Logout ->
            ( Redirect session context
            , Api.logout
            )

        Just (Route.Home activePanel) ->
            case Session.viewer session of
                Nothing ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) Route.Login
                    )

                Just _ ->
                    HomePage.init session context activePanel
                        |> updateWith Home GotHomeMsg currentPage

        Just Route.Login ->
            case Session.viewer session of
                Just _ ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) (Route.Home ActivePanel.None)
                    )

                Nothing ->
                    LoginPage.init session context
                        |> updateWith Login GotLoginMsg currentPage

        Just (Route.Build id) ->
            case Session.viewer session of
                Nothing ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) Route.Login
                    )

                Just _ ->
                    BuildPage.init session context id
                        |> updateWith Build GotBuildMsg currentPage

        Just (Route.Project slug) ->
            case Session.viewer session of
                Nothing ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) Route.Login
                    )

                Just _ ->
                    ProjectPage.init session context slug
                        |> updateWith Project GotProjectMsg currentPage


updatePage : Msg -> Body -> ( Body, Cmd Msg )
updatePage msg page =
    case ( msg, page ) of
        ( Ignored, _ ) ->
            ( page, Cmd.none )

        ( ClickedLink urlRequest, _ ) ->
            case urlRequest of
                Browser.Internal url ->
                    ( page
                    , Nav.pushUrl (Session.navKey (toSession page)) (Url.toString url)
                    )

                Browser.External href ->
                    ( page
                    , Nav.load href
                    )

        ( ChangedUrl url, _ ) ->
            changeRouteTo (Route.fromUrl url) page

        ( ChangedRoute route, _ ) ->
            changeRouteTo route page

        ( WindowResized width height, _ ) ->
            ( updateContext (Context.windowResize { width = width, height = height } (toContext page)) page
            , Cmd.none
            )

        ( GotHomeMsg subMsg, Home home ) ->
            HomePage.update subMsg home
                |> updateWith Home GotHomeMsg page

        ( GotLoginMsg subMsg, Login login ) ->
            LoginPage.update subMsg login
                |> updateWith Login GotLoginMsg page

        ( GotProjectMsg subMsg, Project project ) ->
            ProjectPage.update subMsg project
                |> updateWith Project GotProjectMsg page

        ( GotBuildMsg subMsg, Build build ) ->
            BuildPage.update subMsg build
                |> updateWith Build GotBuildMsg page

        ( UpdateSession task, _ ) ->
            ( page, Task.attempt UpdatedSession task )

        ( UpdatedSession (Ok session), Redirect _ _ ) ->
            ( Redirect session (toContext page)
            , case Session.viewer session of
                Just _ ->
                    Route.replaceUrl (Session.navKey session) (Route.Home ActivePanel.None)

                Nothing ->
                    Route.replaceUrl (Session.navKey session) Route.Login
            )

        ( UpdatedSession (Err _), _ ) ->
            ( page, Cmd.none )

        ( SocketMsg subMsg, _ ) ->
            let
                ( context, socketCmd ) =
                    Context.updateSocket subMsg (toContext page)
            in
            ( updateContext context page
            , Cmd.map SocketMsg socketCmd
            )

        ( _, _ ) ->
            -- Disregard messages that arrived for the wrong page.
            ( page, Cmd.none )


updateContext : Context Msg -> Body -> Body
updateContext context page =
    case page of
        Redirect session _ ->
            Redirect session context

        NotFound session _ ->
            NotFound session context

        Home home ->
            Home { home | context = context }

        Login login ->
            Login { login | context = context }

        Project project ->
            Project { project | context = context }

        Build build ->
            Build { build | context = context }


updateSession : Session -> Body -> Body
updateSession session page =
    case page of
        Redirect _ context ->
            Redirect session context

        NotFound _ context ->
            NotFound session context

        Home home ->
            Home { home | session = session }

        Login login ->
            Login { login | session = session }

        Project project ->
            Project { project | session = session }

        Build build ->
            Build { build | session = session }


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case model of
        ApplicationStarted header page ->
            case msg of
                UpdateLayout newHeader ->
                    ( ApplicationStarted newHeader page
                    , Cmd.none
                    )

                SocketUpdate updateMsg ->
                    let
                        context =
                            toContext page

                        ( session, updatedContext, socketCmd ) =
                            toSession page
                                |> Session.socketUpdate updateMsg SocketUpdate context

                        updatedPage =
                            page
                                |> updateSession session
                                |> updateContext updatedContext
                    in
                    ( ApplicationStarted header updatedPage
                    , Cmd.map SocketMsg socketCmd
                    )

                SetSession session ->
                    ( ApplicationStarted header (updateSession session page)
                    , Cmd.none
                    )

                _ ->
                    updatePage msg page
                        |> Tuple.mapFirst (ApplicationStarted header)

        InitialHTTPRequests _ _ ->
            case msg of
                StartApplication (Ok ( app, pageCmd )) ->
                    let
                        ( context, socketCmd ) =
                            Session.joinChannels (toSession app) SocketUpdate (toContext app)
                    in
                    ( ApplicationStarted Page.initLayout (updateContext context app)
                    , Cmd.batch
                        [ Cmd.map SocketMsg socketCmd
                        , pageCmd
                        ]
                    )

                StartApplication (Err err) ->
                    ( InitError "HTTP error"
                    , Cmd.none
                    )

                _ ->
                    ( model, Cmd.none )

        InitError _ ->
            ( model, Cmd.none )


updateWith : (subPage -> Body) -> (subPageMsg -> Msg) -> Body -> ( subPage, Cmd subPageMsg ) -> ( Body, Cmd Msg )
updateWith toPage toMsg currentPage ( pageModel, pageCmd ) =
    ( toPage pageModel
    , Cmd.map toMsg pageCmd
    )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ layoutSubscriptions model
        , pageSubscriptions model
        , socketSubscriptions model
        , Browser.Events.onResize WindowResized
        ]


socketSubscriptions : Model -> Sub Msg
socketSubscriptions model =
    case model of
        ApplicationStarted _ page ->
            Context.socketSubscriptions SocketMsg (toContext page)

        InitialHTTPRequests context _ ->
            Context.socketSubscriptions SocketMsg context

        InitError _ ->
            Sub.none


layoutSubscriptions : Model -> Sub Msg
layoutSubscriptions model =
    case model of
        ApplicationStarted layout _ ->
            Page.layoutSubscriptions layout UpdateLayout

        InitialHTTPRequests _ _ ->
            Sub.none

        InitError _ ->
            Sub.none


pageSubscriptions : Model -> Sub Msg
pageSubscriptions model =
    case model of
        ApplicationStarted _ page ->
            case page of
                NotFound _ _ ->
                    Sub.none

                Redirect session context ->
                    Session.changes UpdateSession context session

                Home home ->
                    Sub.map GotHomeMsg (HomePage.subscriptions home)

                Login login ->
                    Sub.map GotLoginMsg (LoginPage.subscriptions login)

                Project project ->
                    Sub.map GotProjectMsg (ProjectPage.subscriptions project)

                Build build ->
                    Sub.map GotBuildMsg (BuildPage.subscriptions build)

        InitialHTTPRequests _ _ ->
            Sub.none

        InitError _ ->
            Sub.none



--
---- PROGRAM ----


main : Program Value Model Msg
main =
    Api.application Viewer.decoder Context.start <|
        { onUrlChange = ChangedUrl
        , onUrlRequest = ClickedLink
        , view = view
        , init = init
        , update = update
        , subscriptions = subscriptions
        }
