module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

import Api exposing (Cred)
import Asset
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
    = AddProjectPanel
    | BlankPanel
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
            [ column
                [ width fill
                , height fill
                ]
                (viewPanels (Context.device model.context) (Session.projects model.session))
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


viewPanels : Device -> List Project -> List (Element msg)
viewPanels device projects =
    List.concat
        [ projects, projects, projects ]
        |> List.map ProjectPanel
        |> (::) AddProjectPanel
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
            viewBlankPanel

        AddProjectPanel ->
            viewNewPanel

        ProjectPanel project ->
            viewProjectPanel project


viewBlankPanel : Element msg
viewBlankPanel =
    el
        [ width (fillPortion 2)
        , height (px 150)
        ]
        (text "")


viewNewPanel : Element msg
viewNewPanel =
    column
        [ width (fillPortion 2)
        , height (px 150)
        , Border.width 0
        , Border.color (rgba255 245 245 245 1)
        , Border.rounded 10
        ]
        [ row
            [ height (px 70)
            , width shrink
            , padding 20
            , centerY
            , centerX
            , Background.color (rgba255 245 245 245 1)
            , Border.rounded 360
            ]
            [ image
                [ centerX
                , centerY
                , height (px 30)
                , width (px 30)
                , moveUp 0
                ]
                { src = Asset.src Asset.plus
                , description = "Add project icon"
                }
            , el [ Font.light ] (text "Add project")
            ]
        ]


viewProjectPanel : Project -> Element msg
viewProjectPanel project =
    row
        [ width (fillPortion 2)
        , height (px 150)
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
                    Background.uncropped thumbnail

                Nothing ->
                    Background.color (rgba255 92 184 92 1)
            ]
            (text "")
        , column
            [ width (fillPortion 2)
            , height fill
            , padding 5
            , spacingXY 0 10
            ]
            [ image
                [ width (px 30)
                , height (px 30)
                , alignRight
                ]
                { src = Asset.src Asset.loading
                , description = "Loading spinner"
                }
            , el
                [ alignTop
                , alignLeft
                , Font.extraLight
                , Font.size 20
                , Font.letterSpacing -0.5
                , width fill
                , Font.color (rgba 0 0 0 0.8)
                , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
                , Border.color (rgba255 245 245 245 1)
                , paddingEach { bottom = 5, left = 0, right = 0, top = 0 }
                , clip
                , moveUp 30
                , Font.color (rgba255 92 184 92 1)
                ]
                (text <| Project.name project)
            , paragraph
                [ paddingXY 0 0
                , moveUp 30
                , alignTop
                , alignLeft
                , Font.size 15
                , Font.color (rgba 0 0 0 0.6)
                , Font.medium
                , width fill
                , clipX
                ]
                [ el [ centerX ] (text <| Project.repository project)
                ]
            , paragraph
                [ alignBottom
                , width fill
                , Font.size 13
                , Font.heavy
                , Font.color (rgba 0 0 0 0.6)
                ]
                [ el [ centerX ] (text "Last updated 2 weeks ago")
                ]
            ]
        ]



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
