module Main exposing (Model(..), Msg(..), changeRouteTo, init, main, toSession, update, updateWith, view)

import Api
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
    | Unauthenticated Unauthenticated
    | ApplicationStarted Layout Authenticated


type Unauthenticated
    = LoggingIn (LoginPage.Model Msg)
    | RedirectingToLogin Nav.Key (Context Msg)


type Authenticated
    = Redirect (Session Msg) (Context Msg)
    | NotFound (Session Msg) (Context Msg)
    | Home (HomePage.Model Msg)
    | Project (ProjectPage.Model Msg)
    | Build (BuildPage.Model Msg)


init : Maybe Viewer -> Result Decode.Error (Context Msg) -> Url -> Nav.Key -> ( Model, Cmd Msg )
init maybeViewer contextResult url navKey =
    case contextResult of
        Ok context ->
            ( InitialHTTPRequests context navKey
            , maybeViewer
                |> Session.fromViewer navKey context
                |> Task.map (\s -> changeRouteTo (Route.fromUrl url) (Redirect s context))
                |> Task.attempt StartApplication
            )

        Err error ->
            ( InitError (Decode.errorToString error)
            , Cmd.none
            )



-- VIEW


viewCurrentPage : Layout -> Authenticated -> Document Msg
viewCurrentPage layout currentPage =
    let
        session =
            toSession currentPage

        viewPage page toMsg { title, content } =
            Page.view
                { session = session
                , page = page
                , title = title
                , content = Element.map toMsg content
                , layout = layout
                , updateLayout = UpdateLayout
                , context = toContext currentPage
                }
    in
    case currentPage of
        Redirect _ _ ->
            viewPage Page.Other (always Ignored) BlankPage.view

        NotFound _ _ ->
            viewPage Page.Other (always Ignored) NotFoundPage.view

        Home page ->
            viewPage Page.Home GotHomeMsg (HomePage.view page)

        Project page ->
            viewPage Page.Project GotProjectMsg (ProjectPage.view page)

        Build page ->
            viewPage Page.Build GotBuildMsg (BuildPage.view page)


view : Model -> Document Msg
view model =
    case model of
        ApplicationStarted header currentPage ->
            viewCurrentPage header currentPage

        Unauthenticated (LoggingIn loginModel) ->
            let
                { title, content } =
                    LoginPage.view loginModel
            in
            { title = "Login"
            , body =
                [ layout [ width fill, height fill ] (Element.map GotLoginMsg content)
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

        _ ->
            { title = "Loading"
            , body =
                [ layout
                    [ width fill, height fill ]
                    (el [ centerX, moveDown 100, width shrink, height shrink ] <|
                        Loading.icon { width = 50, height = 50 }
                    )
                ]
            }



---- UPDATE ----


type Msg
    = Ignored
    | StartApplication (Result Session.InitError ( Authenticated, Cmd Msg ))
    | ChangedRoute (Maybe Route)
    | ChangedUrl Url
    | ClickedLink Browser.UrlRequest
    | GotHomeMsg (HomePage.Msg Msg)
    | GotLoginMsg (LoginPage.Msg Msg)
    | GotProjectMsg ProjectPage.Msg
    | GotBuildMsg BuildPage.Msg
    | UpdateSession (Task Session.InitError (Session Msg))
    | UpdatedSession (Result Session.InitError (Session Msg))
    | WindowResized Int Int
    | UpdateLayout Page.Layout
    | SessionSubscription Session.SubscriptionDataMsg


toSession : Authenticated -> Session Msg
toSession page =
    case page of
        Redirect session _ ->
            session

        NotFound session _ ->
            session

        Home home ->
            HomePage.toSession home

        Project project ->
            ProjectPage.toSession project

        Build build ->
            BuildPage.toSession build


toContext : Authenticated -> Context Msg
toContext page =
    case page of
        Redirect _ context ->
            context

        NotFound _ context ->
            context

        Home home ->
            HomePage.toContext home

        Project project ->
            ProjectPage.toContext project

        Build build ->
            BuildPage.toContext build


changeRouteTo : Maybe Route -> Authenticated -> ( Authenticated, Cmd Msg )
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
            HomePage.init session context activePanel
                |> updateWith Home GotHomeMsg currentPage

        Just Route.Login ->
            ( Redirect session context
            , Route.replaceUrl (Session.navKey session) (Route.Home ActivePanel.None)
            )

        Just (Route.Build id) ->
            BuildPage.init session context id
                |> updateWith Build GotBuildMsg currentPage

        Just (Route.Project { slug, maybeAfter }) ->
            ProjectPage.init session context slug maybeAfter
                |> updateWith Project GotProjectMsg currentPage


updatePage : Msg -> Authenticated -> ( Authenticated, Cmd Msg )
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

        ( SessionSubscription subMsg, _ ) ->
            ( updateSession (Session.subscriptionDataUpdate subMsg (toSession page)) page
            , Cmd.none
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

        ( GotProjectMsg subMsg, Project project ) ->
            ProjectPage.update subMsg project
                |> updateWith Project GotProjectMsg page

        ( GotBuildMsg subMsg, Build build ) ->
            BuildPage.update subMsg build
                |> updateWith Build GotBuildMsg page

        --
        --        ( UpdateSession task, _ ) ->
        --            ( page, Task.attempt UpdatedSession task )
        --        ( UpdatedSession (Ok session), Redirect _ _ ) ->
        --            ( Redirect session (toContext page)
        --            , Route.replaceUrl (Session.navKey session) (Route.Home ActivePanel.None)
        --            )
        --
        --        ( UpdatedSession (Err _), _ ) ->
        --            ( page, Cmd.none )
        ( _, _ ) ->
            -- Disregard messages that arrived for the wrong page.
            ( page, Cmd.none )


updateContext : Context Msg -> Authenticated -> Authenticated
updateContext context page =
    case page of
        Redirect session _ ->
            Redirect session context

        NotFound session _ ->
            NotFound session context

        Home home ->
            Home { home | context = context }

        Project project ->
            Project { project | context = context }

        Build build ->
            Build { build | context = context }


updateSession : Session Msg -> Authenticated -> Authenticated
updateSession session page =
    case page of
        Redirect _ context ->
            Redirect session context

        NotFound _ context ->
            NotFound session context

        Home home ->
            Home { home | session = session }

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

                UpdateSession task ->
                    ( model, Task.attempt UpdatedSession task )

                _ ->
                    updatePage msg page
                        |> Tuple.mapFirst (ApplicationStarted header)

        InitialHTTPRequests context navKey ->
            case msg of
                StartApplication (Ok ( app, pageCmd )) ->
                    let
                        ( subscribedSession, subscribeCmd ) =
                            Session.subscribe SessionSubscription (toSession app)
                    in
                    ( ApplicationStarted Page.initLayout (updateSession subscribedSession app)
                    , Cmd.batch
                        [ pageCmd
                        , subscribeCmd
                        ]
                    )

                StartApplication (Err err) ->
                    case err of
                        Session.Unauthenticated ->
                            ( Unauthenticated (RedirectingToLogin navKey context)
                            , Route.replaceUrl navKey Route.Login
                            )

                        _ ->
                            ( InitError "Unexpected error happened!"
                            , Cmd.none
                            )

                _ ->
                    ( model, Cmd.none )

        Unauthenticated (RedirectingToLogin navKey context) ->
            case msg of
                ChangedUrl url ->
                    if Route.fromUrl url == Just Route.Login then
                        let
                            ( loginModel, loginCmd ) =
                                LoginPage.init navKey context
                        in
                        ( Unauthenticated (LoggingIn loginModel)
                        , Cmd.map GotLoginMsg loginCmd
                        )

                    else
                        ( model
                        , Route.replaceUrl navKey Route.Login
                        )

                _ ->
                    ( model
                    , Cmd.none
                    )

        Unauthenticated (LoggingIn loginModel) ->
            case msg of
                UpdatedSession (Ok session) ->
                    let
                        ( subscribedSession, subscribeCmd ) =
                            Session.subscribe SessionSubscription session
                    in
                    ( ApplicationStarted Page.initLayout (Redirect subscribedSession (LoginPage.toContext loginModel))
                    , Cmd.batch
                        [ Route.replaceUrl (Session.navKey subscribedSession) (Route.Home ActivePanel.None)
                        , subscribeCmd
                        ]
                    )

                UpdatedSession (Err _) ->
                    ( model
                    , Route.replaceUrl (LoginPage.toNavKey loginModel) Route.Login
                    )

                UpdateSession task ->
                    ( model, Task.attempt UpdatedSession task )

                GotLoginMsg subMsg ->
                    LoginPage.update subMsg loginModel
                        |> Tuple.mapFirst (LoggingIn >> Unauthenticated)
                        |> Tuple.mapSecond (Cmd.map GotLoginMsg)

                _ ->
                    ( model
                    , Cmd.none
                    )

        InitError _ ->
            ( model, Cmd.none )


updateWith : (subPage -> Authenticated) -> (subPageMsg -> Msg) -> Authenticated -> ( subPage, Cmd subPageMsg ) -> ( Authenticated, Cmd Msg )
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
        , sessionSubscriptions model
        , Browser.Events.onResize WindowResized
        ]


sessionSubscriptions : Model -> Sub Msg
sessionSubscriptions model =
    case model of
        ApplicationStarted _ page ->
            Sub.batch
                [ Session.subscriptions (Task.succeed >> UpdateSession) (toSession page)
                , Session.changes UpdateSession (toContext page) (toSession page |> Session.navKey)
                ]

        Unauthenticated (LoggingIn loginPage) ->
            Session.changes UpdateSession (LoginPage.toContext loginPage) (LoginPage.toNavKey loginPage)

        Unauthenticated (RedirectingToLogin navKey context) ->
            Session.changes UpdateSession context navKey

        _ ->
            Sub.none



--


layoutSubscriptions : Model -> Sub Msg
layoutSubscriptions model =
    case model of
        ApplicationStarted layout _ ->
            Page.layoutSubscriptions layout UpdateLayout

        _ ->
            Sub.none


pageSubscriptions : Model -> Sub Msg
pageSubscriptions model =
    case model of
        ApplicationStarted _ page ->
            case page of
                NotFound _ _ ->
                    Sub.none

                Redirect session context ->
                    Session.changes UpdateSession context (Session.navKey session)

                Home home ->
                    Sub.map GotHomeMsg (HomePage.subscriptions home)

                Project project ->
                    Sub.map GotProjectMsg (ProjectPage.subscriptions project)

                Build build ->
                    Sub.map GotBuildMsg (BuildPage.subscriptions build)

        Unauthenticated _ ->
            Sub.none

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
