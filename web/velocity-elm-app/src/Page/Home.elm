module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

import Api exposing (Cred)
import Browser.Dom as Dom
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Input as Input
import Html.Events exposing (onClick)
import Loading
import Session exposing (Session)
import Task exposing (Task)
import Url.Builder
import Username exposing (Username)



-- MODEL


type alias Model =
    { session : Session
    , context : Context
    }


type Status a
    = Loading
    | LoadingSlowly
    | Loaded a
    | Failed


init : Session -> Context -> ( Model, Cmd Msg )
init session context =
    ( { session = session, context = context }
    , Cmd.batch
        [ Task.perform (\_ -> PassedSlowLoadThreshold) Loading.slowThreshold
        ]
    )



-- VIEW


view : Model -> { title : String, content : Element Msg }
view model =
    { title = "Home"
    , content =
        wrappedRow
            [ width fill
            , height fill
            , padding 10
            , spacing 10
            ]
            [ el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            , el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            , el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            , el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            , el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            , el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color (rgba255 92 184 92 1)
                , alignTop
                ]
                (text "box")
            ]
    }



-- UPDATE


type Msg
    = GotSession Session
    | PassedSlowLoadThreshold


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotSession session ->
            ( { model | session = session }, Cmd.none )

        PassedSlowLoadThreshold ->
            let
                -- If any data is still Loading, change it to LoadingSlowly
                -- so `view` knows to render a spinner.
                model_ =
                    model
            in
            ( model_, Cmd.none )


scrollToTop : Task x ()
scrollToTop =
    Dom.setViewport 0 0
        -- It's not worth showing the user anything special if scrolling fails.
        -- If anything, we'd log this to an error recording service.
        |> Task.onError (\_ -> Task.succeed ())



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Session.changes GotSession (Session.navKey model.session)



-- EXPORT


toSession : Model -> Session
toSession model =
    model.session


toContext : Model -> Context
toContext model =
    model.context
