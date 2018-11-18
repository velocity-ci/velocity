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
import Element.ProjectForm as ProjectForm
import Icon
import Loading
import Page.Home.ActivePanel as ActivePanel exposing (ActivePanel)
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
    , projectForm : Status ProjectForm.State
    }


type Status a
    = NotLoaded
    | Loading
    | LoadingSlowly
    | Loaded a
    | Failed


type Panel
    = AddProjectPanel
    | BlankPanel
    | ProjectPanel Project


init : Session -> Context -> Maybe ActivePanel -> ( Model, Cmd Msg )
init session context maybeActivePanel =
    let
        ( projectForm, projectFormCmd ) =
            case maybeActivePanel of
                Just ActivePanel.NewProjectForm ->
                    ( Loaded ProjectForm.init
                    , Cmd.none
                    )

                Just (ActivePanel.ConfigureProjectForm repository) ->
                    ( Loading
                    , ProjectForm.parseGitUrlCmd repository True
                    )

                _ ->
                    ( NotLoaded
                    , Cmd.none
                    )
    in
    ( { session = session
      , context = context
      , projectForm = projectForm
      }
    , Cmd.batch
        [ Task.perform (\_ -> PassedSlowLoadThreshold) Loading.slowThreshold
        , projectFormCmd
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
            , Background.color Palette.white
            , centerX
            , spacing 20
            ]
            [ viewSubHeader (Context.device model.context)
            , row
                [ width (fill |> maximum 1600)
                , alignRight
                , height fill
                ]
                (viewColumns model (Context.device model.context) (Session.projects model.session))
            ]
    }


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions


viewSubHeader : Device -> Element msg
viewSubHeader device =
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
                    , Background.color Palette.primary5
                    , Border.width 1
                    , Border.rounded 10
                    , Border.color Palette.neutral6
                    , Font.color Palette.primary6
                    , mouseOver [ Background.color Palette.primary5 ]
                    ]
                    (row
                        [ height fill
                        , width fill
                        ]
                        [ Icon.plus { iconOptions | size = 21 }
                        , el [ centerX ] (text "New project")
                        ]
                    )
                    (Route.Home (Just ActivePanel.NewProjectForm))
                , el [ width (fillPortion 1) ] none
                ]

        _ ->
            row
                [ Font.bold
                , Font.size 18
                , width (fill |> maximum 1600)
                , alignRight
                , height shrink
                , paddingXY 15 10
                , Font.color Palette.white
                , Border.widthEach { top = 1, bottom = 1, left = 0, right = 0 }
                , Border.color Palette.neutral6
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
                    , Background.color Palette.primary2
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
                    (Route.Home (Just ActivePanel.NewProjectForm))
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
            2

        ( BigDesktop, _ ) ->
            3


viewColumns : Model -> Device -> List Project -> List (Element Msg)
viewColumns model device projects =
    projects
        |> List.map ProjectPanel
        |> (\projects_ ->
                case model.projectForm of
                    Loaded projectForm ->
                        if ProjectForm.isConfiguring projectForm then
                            AddProjectPanel :: projects_

                        else
                            AddProjectPanel :: projects_

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
                    (List.map (viewPanel device model) panels)
            )


viewPanel : Device -> Model -> Panel -> Element Msg
viewPanel device model panel =
    case panel of
        BlankPanel ->
            viewBlankPanel

        AddProjectPanel ->
            viewProjectFormPanel model.projectForm

        ProjectPanel project ->
            viewProjectPanel project


viewBlankPanel : Element Msg
viewBlankPanel =
    el
        [ width (fillPortion 2)
        , height (px 150)
        ]
        (text "")


viewProjectFormPanel : Status ProjectForm.State -> Element Msg
viewProjectFormPanel projectFormStatus =
    case projectFormStatus of
        Loaded projectForm ->
            column
                [ width (fillPortion 2)
                , Border.width 2
                , Border.color Palette.neutral6
                , Background.color Palette.white
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
                    , inFront
                        (Route.link
                            [ width (px 20)
                            , height (px 20)
                            , Border.width 1
                            , Border.rounded 5
                            , Border.color Palette.neutral4
                            , alignRight
                            , mouseOver
                                [ Background.color Palette.neutral2
                                , Font.color Palette.white
                                ]
                            ]
                            (Icon.x Icon.defaultOptions)
                            (Route.Home Nothing)
                        )
                    ]
                    (text "New project")
                , el
                    [ width (fillPortion 2)
                    , height fill
                    , paddingXY 5 0
                    ]
                    (ProjectForm.view projectForm UpdateProjectForm)
                ]

        _ ->
            none


viewProjectPanel : Project -> Element msg
viewProjectPanel project =
    row
        [ width fill
        , Border.width 1
        , Border.color Palette.primary6
        , Border.rounded 10
        , Background.color Palette.white
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
    | UpdateProjectForm ProjectForm.State (Cmd Msg)
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

        UpdateProjectForm projectForm subCmd ->
            ( { model | projectForm = Loaded projectForm }, subCmd )

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
    Sub.batch
        [ Session.changes UpdateSession (Context.baseUrl model.context) model.session
        , ProjectForm.subscriptions UpdateProjectForm
        ]



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
