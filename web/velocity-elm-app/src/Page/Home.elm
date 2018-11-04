module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

import Api exposing (Cred)
import Browser.Dom as Dom
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
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
    , projects : List Int
    }


type Status a
    = Loading
    | LoadingSlowly
    | Loaded a
    | Failed


init : Session -> Context -> ( Model, Cmd Msg )
init session context =
    ( { session = session
      , context = context
      , projects = List.range 0 20
      }
    , Cmd.batch
        [ Task.perform (\_ -> PassedSlowLoadThreshold) Loading.slowThreshold
        ]
    )



-- VIEW


view : Model -> { title : String, content : Element Msg }
view model =
    { title = "Home"
    , content =
        column
            [ width fill
            , height fill
            , paddingXY 0 10
            , centerX
            ]
            [ el
                [ paddingEach { top = 10, right = 0, bottom = 0, left = 0 }
                , Font.size 16
                , Font.color (rgba255 36 41 46 0.7)
                ]
                (text "Projects")
            , column
                [ width fill
                , height fill
                ]
                (viewBoxRows (Context.device model.context) model.projects)
            ]
    }


splitProjectsToRows : Int -> List a -> List (List a)
splitProjectsToRows i list =
    case List.take i list of
        [] ->
            []

        listHead ->
            listHead :: splitProjectsToRows i (List.drop i list)


rowAmount : Device -> Int
rowAmount device =
    case ( device.class, device.orientation ) of
        ( Phone, Portrait ) ->
            1

        ( Phone, Landscape ) ->
            3

        ( Tablet, Portrait ) ->
            3

        ( Tablet, Landscape ) ->
            6

        ( Desktop, _ ) ->
            7

        ( BigDesktop, _ ) ->
            9


viewBoxRows : Device -> List a -> List (Element msg)
viewBoxRows device projects =
    projects
        |> splitProjectsToRows (rowAmount device)
        |> List.map
            (\i ->
                row
                    [ spacing 20
                    , paddingXY 0 10
                    , width fill
                    , height (fillPortion 1 |> minimum 150 |> maximum 250)
                    ]
                    (List.map (always viewBox) i)
            )


viewBox : Element msg
viewBox =
    el
        [ width (fillPortion 1)
        , height (fillPortion 1 |> minimum 150 |> maximum 250)
        , Border.width 1
        , Border.color (rgba255 92 184 92 1)
        , Border.rounded 10
        ]
        (text "box")



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
