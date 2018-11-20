port module Page.Home exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

{-| The homepage. You can get here via either the / route.
-}

import Api exposing (Cred)
import Array exposing (Array)
import Asset
import Browser.Dom as Dom
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input
import Form.Input as Input
import GitUrl exposing (GitUrl)
import Icon
import Json.Decode as Decode
import Json.Encode as Encode
import Loading
import Page.Home.ActivePanel as ActivePanel exposing (ActivePanel)
import Palette
import Porter
import Project exposing (Project)
import Regex
import Route
import Session exposing (Session)
import Task exposing (Task)
import Url.Builder
import Url.Parser.Query as Query
import Username exposing (Username)
import Validate exposing (ifBlank)



---- Ports


port parseRepository : Encode.Value -> Cmd msg


port parsedRepository : (Decode.Value -> msg) -> Sub msg



---- Model


type alias Model =
    { session : Session
    , context : Context
    , projectFormStatus : ProjectFormStatus
    }


type ProjectFormStatus
    = NotOpen
    | SettingRepository { value : String, dirty : Bool, problems : List String }
    | ConfiguringRepository String GitUrl


init : Session -> Context -> ActivePanel -> ( Model, Cmd Msg )
init session context activePanel =
    ( { session = session
      , context = context
      , projectFormStatus =
            ConfiguringRepository "https://github.com/velocity-ci/velocity.git"
                { protocol = "https"
                , port_ = Nothing
                , resource = "github.com"
                , source = "github.com"
                , owner = "velocity-ci"
                , pathName = "/velocity-ci/velocity.git"
                , fullName = "velocity-ci/velocity"
                , href = "https://github.com/velocity-ci/velocity.git"
                }

      --      , projectFormStatus = activePanelToProjectFormStatus activePanel
      }
    , Cmd.batch
        [ Task.perform (\_ -> PassedSlowLoadThreshold) Loading.slowThreshold
        ]
    )


activePanelToProjectFormStatus : ActivePanel -> ProjectFormStatus
activePanelToProjectFormStatus activePanel =
    case activePanel of
        ActivePanel.ProjectForm ->
            SettingRepository { value = "", dirty = False, problems = validateRepository "" }

        ActivePanel.None ->
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
            [ viewSubHeader (model.projectFormStatus /= NotOpen) device
            , viewPanelGrid device model.projectFormStatus model.session
            ]
    }


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions



-- SubHeader


viewSubHeader : Bool -> Device -> Element msg
viewSubHeader disableButton device =
    case device.class of
        Phone ->
            viewMobileSubHeader { disableButton = disableButton }

        Tablet ->
            viewDesktopSubHeader { disableButton = disableButton }

        Desktop ->
            viewDesktopSubHeader { disableButton = disableButton }

        BigDesktop ->
            viewDesktopSubHeader { disableButton = disableButton }


viewMobileSubHeader : { disableButton : Bool } -> Element msg
viewMobileSubHeader { disableButton } =
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
        [ el [ width fill ] none
        , Button.link (Route.Home ActivePanel.ProjectForm)
            { leftIcon = Nothing
            , rightIcon = Nothing
            , centerLeftIcon = Just Icon.plus
            , centerRightIcon = Nothing
            , size = Button.Large
            , scheme = Button.Primary
            , content = text "New project"
            , widthLength = fillPortion 2
            , disabled = disableButton
            }
        , el [ width fill ] none
        ]


viewDesktopSubHeader : { disableButton : Bool } -> Element msg
viewDesktopSubHeader { disableButton } =
    row
        [ Font.bold
        , Font.size 18
        , width (fill |> maximum 1600)
        , alignRight
        , height shrink
        , paddingXY 20 10
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
        , Button.link (Route.Home ActivePanel.ProjectForm)
            { leftIcon = Nothing
            , rightIcon = Nothing
            , centerLeftIcon = Nothing
            , centerRightIcon = Just Icon.plus
            , size = Button.Medium
            , scheme = Button.Primary
            , content = text "New project"
            , widthLength = shrink
            , disabled = disableButton
            }
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
            , paddingXY 20 0
            , spacingXY 20 0
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
            viewAddProjectPanel repositoryValue

        ProjectFormPanel (ConfiguringRepository repository gitUrl) ->
            viewConfigureProjectPanel repository gitUrl

        ProjectPanel project ->
            viewProjectPanel project



-- Supported panel types


{-| This is a panel that is used to for the the simple form to add a new project, its a bit of a CTA.
-}
viewAddProjectPanel : { value : String, dirty : Bool, problems : List String } -> Element Msg
viewAddProjectPanel repositoryField =
    viewPanelContainer
        [ viewNewProjectPanelHeader
        , row
            [ width fill
            , height fill
            , padding 5
            ]
            [ column [ spacingXY 0 20, width fill ]
                [ column
                    [ spacingXY 0 20
                    , Font.color Palette.neutral3
                    , width fill
                    , Font.size 15
                    , Font.alignLeft
                    , paddingEach { top = 10, left = 0, right = 0, bottom = 0 }
                    ]
                    [ paragraph []
                        [ text "Set up continuous integration or deployment based on a source code repository." ]
                    , paragraph []
                        [ text "This should be a repository with a .velocity.yml file in the root. Check out "
                        , link [ Font.color Palette.primary5 ]
                            { url = "https://google.com", label = text "the documentation" }
                        , text " to find out more."
                        ]
                    ]
                , column [ width fill ]
                    [ viewRepositoryField repositoryField
                    ]
                , row [ width fill, spacingXY 10 0 ]
                    [ el [ width fill ]
                        (Button.link (Route.Home ActivePanel.None)
                            { rightIcon = Nothing
                            , centerRightIcon = Nothing
                            , leftIcon = Nothing
                            , centerLeftIcon = Nothing
                            , content = text "Cancel"
                            , scheme = Button.Transparent
                            , size = Button.Medium
                            , widthLength = fill
                            , disabled = False
                            }
                        )
                    , el [ width (fillPortion 2) ]
                        (Button.button ConfigureRepositoryButtonClicked
                            { rightIcon = Just Icon.arrowRight
                            , centerRightIcon = Nothing
                            , leftIcon = Nothing
                            , centerLeftIcon = Just Icon.settings
                            , content = text "Configure"
                            , scheme = Button.Primary
                            , size = Button.Medium
                            , widthLength = fill
                            , disabled = not (List.isEmpty repositoryField.problems)
                            }
                        )
                    ]
                ]
            ]
        ]


viewConfigureProjectPanel : String -> GitUrl -> Element msg
viewConfigureProjectPanel repository gitUrl =
    let
        infoRow header value =
            row
                [ width fill
                , Font.size 16
                , spacingXY 10 0
                , Background.color Palette.white
                , Border.rounded 10
                , paddingXY 0 10
                ]
                [ el
                    [ width fill
                    , Font.alignRight
                    , Font.color Palette.neutral2
                    ]
                    (text header)
                , el
                    [ width (fillPortion 2)
                    , Font.alignLeft
                    , Font.color Palette.primary2
                    ]
                    (text value)
                ]
    in
    viewPanelContainer
        [ viewNewProjectPanelHeader
        , column [ width fill, spacingXY 0 10 ]
            [ infoRow "Protocol" gitUrl.protocol
            , infoRow "Owner" gitUrl.owner
            , infoRow "Name" gitUrl.fullName
            , infoRow "Source" gitUrl.source
            ]
        ]


viewNewProjectPanelHeader : Element msg
viewNewProjectPanelHeader =
    row
        [ alignTop
        , alignLeft
        , Font.extraLight
        , Font.size 20
        , Font.letterSpacing -0.5
        , width fill
        , Font.color Palette.neutral4
        , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
        , Border.color Palette.primary7
        , paddingEach { top = 5, left = 5, bottom = 10, right = 10 }
        , clip
        , Font.color Palette.primary4

        --- X button
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


viewRepositoryField : { value : String, dirty : Bool, problems : List String } -> Element Msg
viewRepositoryField { value, dirty, problems } =
    Input.text
        { leftIcon = Just Icon.link
        , rightIcon =
            if dirty && List.isEmpty problems then
                Just Icon.check

            else
                Nothing
        , label = Input.labelHidden "Repository URL"
        , placeholder = Just "Repository URL"
        , dirty = dirty
        , value = value
        , problems = problems
        , onChange = EnteredRepositoryUrl
        }


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
                , Font.alignLeft
                ]
                [ paragraph
                    [ alignTop
                    , alignLeft
                    , centerX
                    , Font.extraLight
                    , Font.size 20
                    , Font.letterSpacing -0.5
                    , width fill
                    , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
                    , Border.color Palette.primary7
                    , paddingEach { top = 5, left = 0, bottom = 10, right = 10 }
                    , clip
                    , Font.color Palette.primary4
                    ]
                    [ text <| Project.name project ]
                , column
                    [ paddingEach { bottom = 0, left = 0, right = 0, top = 10 }
                    , width fill
                    , height fill
                    , spacingXY 0 20
                    ]
                    [ paragraph
                        [ Font.size 15
                        , Font.color Palette.neutral3
                        , Font.medium
                        , width fill
                        , clipX
                        ]
                        [ link []
                            { url = Project.repository project
                            , label = text <| Project.repository project
                            }
                        ]
                    , paragraph
                        [ width fill
                        , Font.alignRight
                        , height shrink
                        , alignBottom
                        , Font.size 13
                        , Font.heavy
                        , Font.color Palette.neutral2
                        ]
                        [ text "Last updated 2 weeks ago" ]
                    ]
                ]
    in
    viewPanelContainer
        [ row
            [ width fill, height fill, spacingXY 10 0 ]
            [ thumbnail
            , details
            ]
        ]


viewPanelContainer : List (Element msg) -> Element msg
viewPanelContainer contents =
    column
        [ width fill
        , padding 10
        , Border.width 1
        , Border.color Palette.primary7
        , Border.rounded 10
        , Background.color Palette.neutral7
        ]
        contents



-- VALIDATION


validateRepository : String -> List String
validateRepository repository =
    let
        maybeRegex =
            "(?:git|ssh|https?|git@[-\\w.]+):(\\/\\/)?(.*?)(\\.git)(\\/?|\\#[-\\d\\w._]+?)$"
                |> Regex.fromStringWith { caseInsensitive = False, multiline = False }
    in
    case maybeRegex of
        Just regex ->
            if Regex.contains regex repository then
                []

            else
                [ "Invalid repository" ]

        Nothing ->
            []



-- UPDATE


type Msg
    = UpdateSession (Task Session.InitError Session)
    | UpdatedSession (Result Session.InitError Session)
    | ConfigureRepositoryButtonClicked
    | ParsedRepository (Result Decode.Error { repository : String, gitUrl : GitUrl })
    | EnteredRepositoryUrl String
    | PassedSlowLoadThreshold
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        EnteredRepositoryUrl repositoryUrl ->
            ( { model
                | projectFormStatus =
                    SettingRepository
                        { value = repositoryUrl
                        , dirty = True
                        , problems = validateRepository repositoryUrl
                        }
              }
            , Cmd.none
            )

        UpdateSession task ->
            ( model, Task.attempt UpdatedSession task )

        UpdatedSession (Ok session) ->
            ( { model | session = session }, Cmd.none )

        UpdatedSession (Err _) ->
            ( model, Cmd.none )

        ConfigureRepositoryButtonClicked ->
            ( model
            , case model.projectFormStatus of
                SettingRepository { value, problems } ->
                    if List.isEmpty problems then
                        parseRepository (Encode.string value)

                    else
                        Cmd.none

                _ ->
                    Cmd.none
            )

        ParsedRepository (Ok { repository, gitUrl }) ->
            ( { model | projectFormStatus = ConfiguringRepository repository gitUrl }
            , Cmd.none
            )

        ParsedRepository (Err _) ->
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
        , parsedRepositorySub
        ]


parsedRepositorySub : Sub Msg
parsedRepositorySub =
    let
        decoder =
            Decode.map2
                (\repository gitUrl -> { repository = repository, gitUrl = gitUrl })
                (Decode.field "repository" Decode.string)
                (Decode.field "gitUrl" GitUrl.decoder)
    in
    parsedRepository (Decode.decodeValue decoder >> ParsedRepository)



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
