module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

import Api exposing (Cred)
import Array exposing (Array)
import Asset
import Browser.Dom as Dom
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Icon
import Loading
import Page.Home.ActivePanel exposing (ActivePanel(..))
import Palette
import Project exposing (Project)
import Route
import Session exposing (Session)
import Task exposing (Task)
import Url.Builder
import Url.Parser.Query as Query
import Username exposing (Username)



-- MODEL


type alias Model =
    { session : Session
    , context : Context
    , activePanel : Maybe ActivePanel
    }


type Status a
    = Loading
    | LoadingSlowly
    | Loaded a
    | Failed


type Panel
    = AddProjectPanel Bool
    | BlankPanel
    | ProjectPanel Project


init : Session -> Context -> Maybe ActivePanel -> ( Model, Cmd Msg )
init session context maybeActivePanel =
    ( { session = session
      , context = context
      , activePanel = maybeActivePanel
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

            --            , paddingXY 0 20
            , centerX
            , spacing 20
            ]
            [ viewProjectHeader (Context.device model.context) model.activePanel
            , row
                [ width fill
                , height fill
                ]
                (viewColumns model.activePanel (Context.device model.context) (Session.projects model.session))
            ]
    }


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions


viewProjectHeader : Device -> Maybe ActivePanel -> Element msg
viewProjectHeader device maybeActivePanel =
    case device.class of
        Phone ->
            row
                [ width fill
                , height shrink
                , Background.color Palette.neutral7
                , Font.color Palette.white
                , Border.widthEach { top = 1, bottom = 1, left = 0, right = 0 }
                , Border.color Palette.neutral6
                , paddingXY 0 15
                , Border.shadow
                    { offset = ( 0, 2 )
                    , size = 2
                    , blur = 2
                    , color = Palette.neutral6
                    }
                ]
                [ el [ width (fillPortion 1) ] none
                , Route.link
                    [ width (fillPortion 2)
                    , paddingXY 10 10
                    , Border.rounded 5
                    , Font.size 21
                    , Background.color
                        (case maybeActivePanel of
                            Just NewProjectForm ->
                                Palette.primary4

                            _ ->
                                Palette.primary5
                        )
                    , Border.width 1
                    , Border.rounded 10
                    , Border.color Palette.neutral6
                    , Font.color Palette.neutral7
                    , mouseOver [ Background.color Palette.primary4 ]
                    ]
                    (row
                        [ height fill
                        , width fill
                        ]
                        [ Icon.plus { iconOptions | size = 21 }
                        , el [ centerX ] (text "New project")
                        ]
                    )
                    (Route.Home (Just NewProjectForm))
                , el [ width (fillPortion 1) ] none
                ]

        _ ->
            row
                [ Font.bold
                , Font.size 18
                , width fill
                , height shrink
                , Background.color Palette.neutral7
                , Font.color Palette.white
                , Border.widthEach { top = 1, bottom = 1, left = 0, right = 0 }
                , Border.color Palette.neutral6
                , paddingXY 20 15
                , Border.shadow
                    { offset = ( 0, 2 )
                    , size = 2
                    , blur = 2
                    , color = Palette.neutral6
                    }
                ]
                [ el
                    [ width fill
                    , centerY
                    , Font.color Palette.neutral3
                    ]
                    (el [ alignLeft ] (text "Your projects"))
                , Route.link
                    [ width shrink
                    , paddingXY 10 10
                    , Border.rounded 5
                    , Font.size 16
                    , Background.color
                        (case maybeActivePanel of
                            Just NewProjectForm ->
                                Palette.primary4

                            _ ->
                                Palette.primary5
                        )
                    , Border.width 1
                    , Border.rounded 10
                    , Border.color Palette.neutral6
                    , Font.color Palette.neutral7
                    , mouseOver [ Background.color Palette.primary4 ]
                    ]
                    (row
                        [ height fill
                        , width fill
                        ]
                        [ Icon.plus { iconOptions | size = 16 }
                        , el [ centerX ] (text "New project")
                        ]
                    )
                    (Route.Home (Just NewProjectForm))
                ]


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


viewColumns : Maybe ActivePanel -> Device -> List Project -> List (Element Msg)
viewColumns maybeActivePanel device projects =
    projects
        |> List.map ProjectPanel
        |> (\projects_ ->
                case maybeActivePanel of
                    Just NewProjectForm ->
                        AddProjectPanel True :: projects_

                    _ ->
                        projects_
           )
        |> List.indexedMap Tuple.pair
        |> List.foldl
            (\( i, panel ) columns ->
                let
                    columnIndex =
                        remainderBy (rowAmount device) i
                in
                case Array.get columnIndex columns of
                    Just columnItems ->
                        Array.set columnIndex (List.append columnItems [ panel ]) columns

                    Nothing ->
                        columns
            )
            (List.range 1 (rowAmount device)
                |> List.map (always [])
                |> Array.fromList
            )
        |> Array.toList
        |> List.map
            (\panels ->
                column
                    [ spacing 20
                    , paddingXY 10 0
                    , width fill
                    , height fill
                    ]
                    (List.map (viewPanel device) panels)
            )


viewPanel : Device -> Panel -> Element Msg
viewPanel device panel =
    case panel of
        BlankPanel ->
            viewBlankPanel

        AddProjectPanel open ->
            viewProjectFormPanel

        ProjectPanel project ->
            viewProjectPanel project


viewBlankPanel : Element Msg
viewBlankPanel =
    el
        [ width (fillPortion 2)
        , height (px 150)
        ]
        (text "")


viewProjectFormPanel : Element Msg
viewProjectFormPanel =
    column
        [ width (fillPortion 2)
        , Border.width 2
        , Border.color (rgba255 245 245 245 1)
        , Border.rounded 10
        , Font.size 14
        , padding 10
        , spacingXY 0 20
        ]
        [ el
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
            , Font.color Palette.primary4
            ]
            (text "New project")
        , column
            [ spacingXY 0 10 ]
            [ paragraph [] [ text "Set up continuous integration or deployment based on a source code repository." ]
            , paragraph []
                [ text "This should be a repository with a .velocity.yml file in the root. Check out "
                , link [ Font.color Palette.primary5 ] { url = "https://google.com", label = text "the documentation" }
                , text " to find out more."
                ]
            ]
        , column
            [ width (fillPortion 2)
            , height fill
            , padding 5
            , spacingXY 0 20
            ]
            [ row [ width fill ]
                [ Input.text [ height (px 30) ]
                    { onChange = always NoOp
                    , placeholder = Nothing
                    , text = ""
                    , label = Input.labelAbove [ alignLeft ] (text "Project name")
                    }
                ]
            , row [ width fill ]
                [ Input.text [ height (px 30) ]
                    { onChange = always NoOp
                    , placeholder = Nothing
                    , text = ""
                    , label = Input.labelAbove [ alignLeft ] (text "Repository URL")
                    }
                ]
            , row
                [ width fill
                , paddingEach { top = 10, left = 0, right = 0, bottom = 0 }
                , spacing 10
                ]
                [ Route.link
                    [ width (fillPortion 1)
                    , height (px 35)
                    , Border.width 1
                    , Border.rounded 5
                    , Border.color Palette.neutral4
                    , alignBottom
                    , mouseOver
                        [ Background.color Palette.neutral2
                        , Font.color Palette.white
                        ]
                    ]
                    (el [ centerY, centerX ] (text "Cancel"))
                    (Route.Home Nothing)
                , Input.button
                    [ width (fillPortion 2)
                    , height (px 35)
                    , Border.width 1
                    , Border.rounded 5
                    , Border.color Palette.primary4
                    , Font.color Palette.primary4
                    , alignBottom
                    , mouseOver
                        [ Background.color Palette.primary4
                        , Font.color Palette.white
                        ]
                    ]
                    { onPress = Just NoOp
                    , label = text "Add"
                    }
                ]
            ]
        ]


viewProjectPanel : Project -> Element msg
viewProjectPanel project =
    row
        [ width fill
        , Border.width 2
        , Border.color Palette.neutral7
        , Border.rounded 10
        , pointer
        , mouseOver [ Background.color Palette.primary7 ]
        ]
        [ el
            [ width (fillPortion 1)
            , height fill
            , padding 10
            ]
            (case Project.thumbnailSrc project of
                Just thumbnail ->
                    el
                        [ width fill
                        , height fill
                        , Background.image thumbnail
                        , Border.width 1
                        , Border.color Palette.neutral5
                        , Border.rounded 10
                        ]
                        (text "")

                Nothing ->
                    text ""
            )
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
                , transparent True
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
                , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
                , Border.color Palette.primary7
                , paddingEach { bottom = 5, left = 0, right = 0, top = 0 }
                , clip
                , moveUp 30
                , Font.color Palette.primary4
                ]
                (text <| Project.name project)
            , paragraph
                [ paddingXY 0 0
                , moveUp 30
                , alignTop
                , alignLeft
                , Font.size 15
                , Font.color Palette.neutral3
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
                , Font.color Palette.neutral2
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
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

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



-- PARSER


activePanelQueryParser : String -> Query.Parser (Maybe ActivePanel)
activePanelQueryParser key =
    Query.custom key <|
        \stringList ->
            case stringList of
                [ "new-project" ] ->
                    Just NewProjectForm

                _ ->
                    Nothing



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



-- UTIL


viewIf : Bool -> Element msg -> Element msg
viewIf condition content =
    if condition then
        content

    else
        none
