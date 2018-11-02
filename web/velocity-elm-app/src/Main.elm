module Main exposing (Model(..), Msg(..), changeRouteTo, init, main, toSession, update, updateWith, view)

import Api
import Api.Endpoint as Endpoint exposing (Endpoint)
import Browser exposing (Document)
import Browser.Navigation as Nav
import Context exposing (Context)
import Html exposing (Html, div, h1, img, text)
import Html.Attributes exposing (src)
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
    | ApplicationStarted Application


type Application
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


viewCurrentPage : Application -> Document Msg
viewCurrentPage currentPage =
    let
        viewPage page toMsg config =
            let
                { title, body } =
                    Page.view (Session.viewer (toSession currentPage)) page config
            in
            { title = title
            , body = List.map (Html.map toMsg) body
            }
    in
    case currentPage of
        Redirect _ _ ->
            viewPage Page.Other (\_ -> Ignored) Blank.view

        NotFound _ _ ->
            viewPage Page.Other (\_ -> Ignored) NotFound.view

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
                [ Html.h1 [] [ text "Error Starting Application" ]
                , Html.pre [] [ text error ]
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


toSession : Application -> Session
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


toContext : Application -> Context
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


changeRouteTo : Maybe Route -> Application -> ( Application, Cmd Msg )
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
            Login.init session context
                |> updateWith Login GotLoginMsg currentPage


updatePage : Msg -> Application -> ( Application, Cmd Msg )
updatePage msg page =
    case ( msg, page ) of
        ( Ignored, _ ) ->
            ( page, Cmd.none )

        ( ClickedLink urlRequest, _ ) ->
            case urlRequest of
                Browser.Internal url ->
                    case url.fragment of
                        Nothing ->
                            -- If we got a link that didn't include a fragment,
                            -- it's from one of those (href "") attributes that
                            -- we have to include to make the RealWorld CSS work.
                            --
                            -- In an application doing path routing instead of
                            -- fragment-based routing, this entire
                            -- `case url.fragment of` expression this comment
                            -- is inside would be unnecessary.
                            ( page, Cmd.none )

                        Just _ ->
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


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case model of
        ApplicationStarted page ->
            updatePage msg page
                |> Tuple.mapFirst ApplicationStarted

        InitError _ ->
            ( model, Cmd.none )


updateWith : (subPage -> Application) -> (subPageMsg -> Msg) -> Application -> ( subPage, Cmd subPageMsg ) -> ( Application, Cmd Msg )
updateWith toPage toMsg currentPage ( pageModel, pageCmd ) =
    ( toPage pageModel
    , Cmd.map toMsg pageCmd
    )



---- PROGRAM ----


main : Program Value Model Msg
main =
    Api.application Viewer.decoder
        Context.fromBaseUrl
        { onUrlChange = ChangedUrl
        , onUrlRequest = ClickedLink
        , view = view
        , init = init
        , update = update
        , subscriptions = always Sub.none
        }
