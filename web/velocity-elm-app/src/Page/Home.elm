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
import Project exposing (Project)
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


type Panel
    = BlankPanel
    | ProjectPanel Project


init : Session -> Context -> ( Model, Cmd Msg )
init session context =
    ( { session = session
      , context = context
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
                (viewBoxRows (Context.device model.context) (Session.projects model.session))
            ]
    }


splitProjectsToRows : Int -> List Panel -> List (List Panel)
splitProjectsToRows i list =
    case List.take i list of
        [] ->
            []

        listHead ->
            let
                head =
                    if List.length listHead < i then
                        List.range 1 (i - List.length listHead)
                            |> List.map (always BlankPanel)
                            |> List.append listHead

                    else
                        listHead
            in
            head :: splitProjectsToRows i (List.drop i list)


rowAmount : Device -> Int
rowAmount device =
    case ( device.class, device.orientation ) of
        ( Phone, Portrait ) ->
            1

        ( Phone, Landscape ) ->
            2

        ( Tablet, Portrait ) ->
            2

        ( Tablet, Landscape ) ->
            2

        ( Desktop, _ ) ->
            3

        ( BigDesktop, _ ) ->
            3


viewBoxRows : Device -> List Project -> List (Element msg)
viewBoxRows device projects =
    projects
        |> List.map ProjectPanel
        |> splitProjectsToRows (rowAmount device)
        |> List.map
            (\i ->
                row
                    [ spacing 20
                    , paddingXY 0 10
                    , width fill
                    , height (fillPortion 1 |> minimum 150 |> maximum 250)
                    ]
                    (List.map viewPanel i)
            )


viewPanel : Panel -> Element msg
viewPanel panel =
    case panel of
        BlankPanel ->
            el
                [ width (fillPortion 1)
                , height (fillPortion 1 |> minimum 150 |> maximum 250)
                ]
                (text "")

        ProjectPanel project ->
            row
                [ width (fillPortion 1)
                , height (fillPortion 1 |> minimum 150 |> maximum 250)
                , Border.width 2
                , Border.color (rgba255 245 245 245 1)
                , Border.rounded 10
                , mouseOver
                    [ Background.gradient
                        { angle = 90
                        , steps =
                            [ rgba255 0 0 0 0
                            , rgba255 0 0 0 0
                            , rgba255 0 0 0 0
                            , rgba255 245 245 245 1
                            ]
                        }
                    ]
                ]
                [ el
                    [ width (fillPortion 1)
                    , height fill
                    , Border.rounded 25
                    , case Project.thumbnailSrc project of
                        Just thumbnail ->
                            Background.image thumbnail

                        Nothing ->
                            Background.color (rgba255 92 184 92 1)
                    ]
                    (text "")
                , column
                    [ width (fillPortion 2)
                    , height fill
                    , padding 10
                    , spaceEvenly
                    ]
                    [ el
                        [ alignTop
                        , alignLeft
                        , Font.medium
                        , Font.size 20
                        , Font.letterSpacing -0.5
                        , width fill
                        , Font.color (rgba 0 0 0 0.8)
                        , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
                        , Border.color (rgba255 245 245 245 1)
                        , paddingEach { bottom = 5, left = 0, right = 0, top = 0 }
                        , clip
                        ]
                        (text <| Project.name project)
                    , paragraph
                        [ paddingXY 0 10
                        , alignTop
                        , alignLeft
                        , Font.size 15
                        , Font.light
                        , width fill
                        , clipX
                        ]
                        [ text <| Project.repository project
                        ]
                    , paragraph
                        [ alignBottom
                        , alignLeft
                        , Font.size 13
                        , Font.light
                        ]
                        [ text "Last updated 2 weeks ago"
                        ]
                    ]
                ]



--
--
--viewProjectThumbnail : Project -> Element msg
--viewProjectThumbnail project =
--    case project.logo of
--        Just logo ->
-- UPDATE


type Msg
    = UpdateSession (Task Session.InitError Session)
    | UpdatedSession (Result Session.InitError Session)
    | PassedSlowLoadThreshold


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UpdateSession task ->
            ( model, Task.attempt UpdatedSession task )

        UpdatedSession (Ok session) ->
            ( { model | session = session }, Cmd.none )

        UpdatedSession (Err _) ->
            ( model, Cmd.none )

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
    Session.changes UpdateSession (Context.baseUrl model.context) model.session



-- EXPORT


toSession : Model -> Session
toSession model =
    model.session


toContext : Model -> Context
toContext model =
    model.context
