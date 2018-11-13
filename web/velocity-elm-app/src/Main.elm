module Main exposing (Model(..), Msg(..), changeRouteTo, init, main, toSession, update, updateWith, view)

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
import Page exposing (Header)
import Page.Blank as Blank
import Page.Home as Home
import Page.Login as Login
import Page.NotFound as NotFound
import Route exposing (Route)
import Session exposing (Session)
import Task exposing (Task)
import Url exposing (Url)
import Viewer exposing (Viewer)



---- MODEL ----


type Model
    = InitError String
    | Initialising Context Nav.Key
    | ApplicationStarted Header Body


type Body
    = Redirect Session Context
    | NotFound Session Context
    | Home Home.Model
    | Login Login.Model


init : Maybe Viewer -> Result Decode.Error Context -> Url -> Nav.Key -> ( Model, Cmd Msg )
init maybeViewer contextResult url navKey =
    case contextResult of
        Ok context ->
            ( Initialising context navKey
            , Session.fromViewer navKey (Context.baseUrl context) maybeViewer
                |> Task.map (\session -> changeRouteTo (Route.fromUrl url) (Redirect session context))
                |> Task.attempt StartApplication
            )

        Err error ->
            ( InitError (Decode.errorToString error), Cmd.none )



-- VIEW


viewCurrentPage : Header -> Body -> Document Msg
viewCurrentPage header currentPage =
    let
        viewPage page toMsg { title, content } =
            Page.view
                { viewer = Session.viewer (toSession currentPage)
                , page = page
                , title = title
                , content = content
                , toMsg = toMsg
                , header = header
                , updateHeader = UpdateHeader
                , context = toContext currentPage
                }
    in
    case currentPage of
        Redirect _ _ ->
            viewPage Page.Other (always Ignored) Blank.view

        NotFound _ _ ->
            viewPage Page.Other (always Ignored) NotFound.view

        Home home ->
            viewPage Page.Home GotHomeMsg (Home.view home)

        Login login ->
            viewPage Page.Login GotLoginMsg (Login.view login)


view : Model -> Document Msg
view model =
    case model of
        ApplicationStarted header currentPage ->
            viewCurrentPage header currentPage

        Initialising _ _ ->
            { title = "Loading"
            , body = [ layout [] (text "Loading") ]
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
    | GotHomeMsg Home.Msg
    | GotLoginMsg Login.Msg
    | UpdateSession (Task Session.InitError Session)
    | UpdatedSession (Result Session.InitError Session)
    | WindowResized Int Int
    | UpdateHeader Page.Header



--    | UserMenuClicked


toSession : Body -> Session
toSession page =
    case page of
        Redirect session _ ->
            session

        NotFound session _ ->
            session

        Home home ->
            Home.toSession home

        Login login ->
            Login.toSession login


toContext : Body -> Context
toContext page =
    case page of
        Redirect _ context ->
            context

        NotFound _ context ->
            context

        Home home ->
            Home.toContext home

        Login login ->
            Login.toContext login


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
            , Route.replaceUrl (Session.navKey session) (Route.Home Nothing)
            )

        Just Route.Logout ->
            ( Redirect session context
            , Api.logout
            )

        Just (Route.Home maybeActivePanel) ->
            case Session.viewer session of
                Nothing ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) Route.Login
                    )

                Just _ ->
                    Home.init session context maybeActivePanel
                        |> updateWith Home GotHomeMsg currentPage

        Just Route.Login ->
            case Session.viewer session of
                Just _ ->
                    ( Redirect session context
                    , Route.replaceUrl (Session.navKey session) (Route.Home Nothing)
                    )

                Nothing ->
                    Login.init session context
                        |> updateWith Login GotLoginMsg currentPage


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
            Home.update subMsg home
                |> updateWith Home GotHomeMsg page

        ( GotLoginMsg subMsg, Login login ) ->
            Login.update subMsg login
                |> updateWith Login GotLoginMsg page

        ( UpdateSession task, _ ) ->
            ( page, Task.attempt UpdatedSession task )

        ( UpdatedSession (Ok session), Redirect _ _ ) ->
            ( Redirect session (toContext page)
            , case Session.viewer session of
                Just _ ->
                    Route.replaceUrl (Session.navKey session) (Route.Home Nothing)

                Nothing ->
                    Route.replaceUrl (Session.navKey session) Route.Login
            )

        ( UpdatedSession (Err _), _ ) ->
            ( page, Cmd.none )

        ( _, _ ) ->
            -- Disregard messages that arrived for the wrong page.
            ( page, Cmd.none )


updateContext : Context -> Body -> Body
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


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case model of
        ApplicationStarted header page ->
            case msg of
                StartApplication (Ok ( app, cmd )) ->
                    ( ApplicationStarted header app
                    , cmd
                    )

                StartApplication (Err err) ->
                    ( InitError (Session.errorToString err)
                    , Cmd.none
                    )

                UpdateHeader newHeader ->
                    ( ApplicationStarted newHeader page
                    , Cmd.none
                    )

                _ ->
                    updatePage msg page
                        |> Tuple.mapFirst (ApplicationStarted header)

        Initialising _ _ ->
            case msg of
                StartApplication (Ok ( app, cmd )) ->
                    ( ApplicationStarted Page.initHeader app
                    , cmd
                    )

                StartApplication (Err err) ->
                    ( InitError (Session.errorToString err)
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
        [ headerSubscriptions model
        , pageSubscriptions model
        , Browser.Events.onResize WindowResized
        ]


headerSubscriptions : Model -> Sub Msg
headerSubscriptions model =
    case model of
        ApplicationStarted header _ ->
            Page.headerSubscriptions header UpdateHeader

        Initialising _ _ ->
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
                    Session.changes UpdateSession (Context.baseUrl context) session

                Home home ->
                    Sub.map GotHomeMsg (Home.subscriptions home)

                Login login ->
                    Sub.map GotLoginMsg (Login.subscriptions login)

        Initialising _ _ ->
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
