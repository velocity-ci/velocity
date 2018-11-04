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
import Page
import Page.Blank as Blank
import Page.Home as Home
import Page.Login as Login
import Page.NotFound as NotFound
import Route exposing (Route)
import Session exposing (Session)
import Url exposing (Url)
import Viewer exposing (Viewer)



---- MODEL ----


type Model
    = InitError String
    | ApplicationStarted App


type App
    = Redirect Session Context
    | NotFound Session Context
    | Home Home.Model
    | Login Login.Model


init : Maybe Viewer -> Result Decode.Error Context -> Url -> Nav.Key -> ( Model, Cmd Msg )
init maybeViewer contextResult url navKey =
    case contextResult of
        Ok context ->
            changeRouteTo (Route.fromUrl url) (Redirect (Session.fromViewer navKey maybeViewer) context)
                |> Tuple.mapFirst ApplicationStarted

        Err error ->
            ( InitError (Decode.errorToString error), Cmd.none )



-- VIEW


viewCurrentPage : App -> Document Msg
viewCurrentPage currentPage =
    let
        viewPage page toMsg config =
            let
                { title, body } =
                    Page.view (Session.viewer (toSession currentPage)) page config toMsg
            in
            { title = title
            , body = body
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
        ApplicationStarted currentPage ->
            viewCurrentPage currentPage

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
    | ChangedRoute (Maybe Route)
    | ChangedUrl Url
    | ClickedLink Browser.UrlRequest
    | GotHomeMsg Home.Msg
    | GotLoginMsg Login.Msg
    | GotSession Session
    | WindowResized Int Int


toSession : App -> Session
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


toContext : App -> Context
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


changeRouteTo : Maybe Route -> App -> ( App, Cmd Msg )
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
            ( currentPage, Route.replaceUrl (Session.navKey session) Route.Home )

        Just Route.Logout ->
            ( currentPage, Api.logout )

        Just Route.Home ->
            Home.init session context
                |> updateWith Home GotHomeMsg currentPage

        Just Route.Login ->
            case Session.viewer session of
                Just _ ->
                    changeRouteTo (Just Route.Home) currentPage

                _ ->
                    Login.init session context
                        |> updateWith Login GotLoginMsg currentPage


updatePage : Msg -> App -> ( App, Cmd Msg )
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

        ( GotSession session, Redirect _ _ ) ->
            ( Redirect session (toContext page)
            , Route.replaceUrl (Session.navKey session) Route.Home
            )

        ( _, _ ) ->
            -- Disregard messages that arrived for the wrong page.
            ( page, Cmd.none )


updateContext : Context -> App -> App
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
        ApplicationStarted page ->
            updatePage msg page
                |> Tuple.mapFirst ApplicationStarted

        InitError _ ->
            ( model, Cmd.none )


updateWith : (subPage -> App) -> (subPageMsg -> Msg) -> App -> ( subPage, Cmd subPageMsg ) -> ( App, Cmd Msg )
updateWith toPage toMsg currentPage ( pageModel, pageCmd ) =
    ( toPage pageModel
    , Cmd.map toMsg pageCmd
    )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ pageSubscriptions model
        , Browser.Events.onResize WindowResized
        ]


pageSubscriptions : Model -> Sub Msg
pageSubscriptions model =
    case model of
        ApplicationStarted page ->
            case page of
                NotFound _ _ ->
                    Sub.none

                Redirect _ _ ->
                    Session.changes GotSession (Session.navKey (toSession page))

                Home home ->
                    Sub.map GotHomeMsg (Home.subscriptions home)

                Login login ->
                    Sub.map GotLoginMsg (Login.subscriptions login)

        InitError _ ->
            Sub.none



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
