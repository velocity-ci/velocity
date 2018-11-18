module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

--import Element.ProjectForm as ProjectForm

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
import Element.Input
import Form.Input as Input
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



---- Model


type alias Model =
    { session : Session
    , context : Context
    , projectFormStatus : ProjectFormStatus
    }


type ProjectFormStatus
    = NotOpen
    | SettingRepository { value : String, dirty : Bool, problems : List String }
    | ConfiguringRepository String


init : Session -> Context -> ActivePanel -> ( Model, Cmd Msg )
init session context activePanel =
    ( { session = session
      , context = context
      , projectFormStatus = activePanelToProjectFormStatus activePanel
      }
    , Cmd.batch
        [ Task.perform (\_ -> PassedSlowLoadThreshold) Loading.slowThreshold
        ]
    )


activePanelToProjectFormStatus : ActivePanel -> ProjectFormStatus
activePanelToProjectFormStatus activePanel =
    case activePanel of
        ActivePanel.ProjectForm ->
            SettingRepository { value = "", dirty = False, problems = [] }

        _ ->
            NotOpen



---- View


view : Model -> { title : String, content : Element Msg }
view model =
    let
        device =
            Context.device model.context
    in
    { title = "Home"
    , content =
        column
            [ width fill
            , height fill
            , Background.color Palette.white
            , centerX
            , spacing 20
            ]
            [ viewSubHeader device
            , viewPanelGrid device model.projectFormStatus model.session
            ]
    }


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions



-- SubHeader


viewSubHeader : Device -> Element msg
viewSubHeader device =
    case device.class of
        Phone ->
            viewMobileSubHeader

        Tablet ->
            viewDesktopSubHeader

        Desktop ->
            viewDesktopSubHeader

        BigDesktop ->
            viewDesktopSubHeader


viewMobileSubHeader : Element msg
viewMobileSubHeader =
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
            (viewNewProjectButton 21)
            (Route.Home ActivePanel.ProjectForm)
        , el [ width (fillPortion 1) ] none
        ]


viewDesktopSubHeader : Element msg
viewDesktopSubHeader =
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
            (viewNewProjectButton 16)
            (Route.Home ActivePanel.ProjectForm)
        ]


viewNewProjectButton : Float -> Element msg
viewNewProjectButton iconSize =
    row
        [ height fill
        , width fill
        ]
        [ Icon.plus { iconOptions | size = iconSize }
        , el [ centerX ] (text "New project")
        ]



-- Panel grid


type Panel
    = ProjectPanel Project
    | ProjectFormPanel ProjectFormStatus


colAmount : Device -> Int
colAmount device =
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


{-| Splits up the all of the data in the model in to "panels" and inserts them into a grid, of size specified by
the device. Finally wraps all of this in a row with a max-width
-}
viewPanelGrid : Device -> ProjectFormStatus -> Session -> Element Msg
viewPanelGrid device projectFormStatus session =
    toPanels projectFormStatus session
        |> List.indexedMap Tuple.pair
        |> List.foldl
            (\( i, panel ) columns ->
                let
                    columnIndex =
                        remainderBy (colAmount device) i
                in
                case Array.get columnIndex columns of
                    Just columnItems ->
                        Array.set columnIndex (List.append columnItems [ panel ]) columns

                    Nothing ->
                        columns
            )
            (List.range 1 (colAmount device)
                |> List.map (always [])
                |> Array.fromList
            )
        |> Array.toList
        |> List.map
            (\panels ->
                column
                    [ spacing 20
                    , width fill
                    , height fill
                    ]
                    (List.map (viewPanel device) panels)
            )
        |> row
            [ width (fill |> maximum 1600)
            , paddingXY 10 0
            , spacingXY 10 0
            , alignRight
            , height fill
            ]


toPanels : ProjectFormStatus -> Session -> List Panel
toPanels projectFormStatus session =
    let
        projectPanels =
            session
                |> Session.projects
                |> List.map ProjectPanel
    in
    if projectFormStatus /= NotOpen then
        ProjectFormPanel projectFormStatus :: projectPanels

    else
        projectPanels


viewPanel : Device -> Panel -> Element Msg
viewPanel device panel =
    case panel of
        ProjectFormPanel NotOpen ->
            none

        ProjectFormPanel (SettingRepository repositoryValue) ->
            viewProjectFormSettingRepositoryPanel repositoryValue

        ProjectFormPanel (ConfiguringRepository _) ->
            text "helloworld"

        ProjectPanel project ->
            viewProjectPanel project



-- Supported panel types


viewProjectFormSettingRepositoryPanel : { value : String, dirty : Bool, problems : List String } -> Element Msg
viewProjectFormSettingRepositoryPanel { value, dirty, problems } =
    viewPanelContainer
        [ row
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
                    (Route.Home ActivePanel.None)
                )
            ]
            [ text "New project" ]
        , row
            [ width fill
            , height fill
            , paddingXY 5 0
            ]
            [ column [ spacingXY 0 20, width fill ]
                [ viewHelpText
                , column [ width fill ]
                    [--             viewUrlField url maybeGitUrl (\newUrl parseCmd -> updateMsg (CheckingUrl newUrl) parseCmd)
                    ]
                , row [ width fill ]
                    [ el [ width fill ] none
                    , el [ width (fillPortion 3) ] (text "next b")
                    ]
                ]
            ]
        ]


viewHelpText : Element msg
viewHelpText =
    column [ spacingXY 0 20, Font.color Palette.neutral3, width fill ]
        [ column [ alignLeft ]
            [ paragraph [ alignLeft ] [ text "Set up continuous integration or deployment based on a source code repository." ]
            ]
        , column
            []
            [ paragraph []
                [ text "This should be a repository with a .velocity.yml file in the root. Check out "
                , link [ Font.color Palette.primary5 ] { url = "https://google.com", label = text "the documentation" }
                , text " to find out more."
                ]
            ]
        ]



--viewRepositoryField : { value : String, dirty : Bool, problems : List String } -> Element Msg
--viewRepositoryField { value, dirty, problems } =
--
--viewProjectFormPanel : Status ProjectForm.State -> Element Msg
--viewProjectFormPanel projectFormStatus =
--    case projectFormStatus of
--        Loaded projectForm ->
--            column
--                [ width (fillPortion 2)
--                , Border.width 2
--                , Border.color Palette.neutral6
--                , Background.color Palette.white
--                , Border.rounded 10
--                , Font.size 14
--                , padding 10
--                , spacingXY 0 20
--                ]
--                [ el
--                    [ alignTop
--                    , alignLeft
--                    , Font.extraLight
--                    , Font.size 20
--                    , Font.letterSpacing -0.5
--                    , width fill
--                    , Font.color (rgba 0 0 0 0.8)
--                    , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
--                    , Border.color (rgba255 245 245 245 1)
--                    , paddingEach { bottom = 5, left = 0, right = 0, top = 0 }
--                    , clip
--                    , Font.color Palette.primary4
--                    , inFront
--                        (Route.link
--                            [ width (px 20)
--                            , height (px 20)
--                            , Border.width 1
--                            , Border.rounded 5
--                            , Border.color Palette.neutral4
--                            , alignRight
--                            , mouseOver
--                                [ Background.color Palette.neutral2
--                                , Font.color Palette.white
--                                ]
--                            ]
--                            (Icon.x Icon.defaultOptions)
--                            (Route.Home Nothing)
--                        )
--                    ]
--                    (text "New project")
--                , el
--                    [ width (fillPortion 2)
--                    , height fill
--                    , paddingXY 5 0
--                    ]
--                    (ProjectForm.view projectForm UpdateProjectForm)
--                ]
--
--        _ ->
--            none


viewProjectPanel : Project -> Element msg
viewProjectPanel project =
    let
        thumbnail =
            column
                [ width (fillPortion 1)
                , height fill
                ]
                [ Project.thumbnail project ]

        details =
            column
                [ width (fillPortion 2)
                , height fill
                , padding 5
                , spacingXY 0 10
                ]
                [ el
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
                    , Font.color Palette.primary4
                    ]
                    (text <| Project.name project)
                , paragraph
                    [ paddingXY 0 0
                    , centerY
                    , alignLeft
                    , height fill
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
    in
    viewPanelContainer
        [ row
            [ width fill, height fill ]
            [ thumbnail
            , details
            ]
        ]


viewPanelContainer : List (Element msg) -> Element msg
viewPanelContainer contents =
    column
        [ width fill
        , height (px 130)
        , padding 10
        , spacingXY 5 0
        , Border.width 1
        , Border.color Palette.primary6
        , Border.rounded 10
        , Background.color Palette.white
        , pointer
        , mouseOver [ Background.color Palette.primary7 ]
        ]
        contents



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



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Session.changes UpdateSession (Context.baseUrl model.context) model.session
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
