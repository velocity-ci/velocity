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
import Http
import Icon
import Json.Decode as Decode
import Json.Encode as Encode
import KnownHost exposing (KnownHost)
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


type alias Model msg =
    { session : Session
    , context : Context msg
    , projectFormStatus : ProjectFormStatus
    }


type ProjectFormStatus
    = NotOpen
    | SettingRepository { value : String, dirty : Bool, problems : List String }
    | AddingKnownHost { gitUrl : GitUrl, publicKey : String }
    | ConfiguringRepository { gitUrl : GitUrl, projectName : String }


init : Session -> Context msg -> ActivePanel -> ( Model msg, Cmd Msg )
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
            SettingRepository { value = "", dirty = False, problems = validateRepository "" }

        ActivePanel.None ->
            NotOpen



---- View


view : Model msg -> { title : String, content : Element Msg }
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
        , Background.color Palette.white
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
            , heightLength = px 45
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
        , height (px 65)
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
            , heightLength = px 45
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
            , Background.color Palette.neutral7
            , paddingXY 20 20
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

        ProjectFormPanel (SettingRepository repositoryField) ->
            viewAddProjectPanel repositoryField

        ProjectFormPanel (AddingKnownHost configurationValues) ->
            viewAddKnownHostPanel configurationValues

        ProjectFormPanel (ConfiguringRepository configurationValues) ->
            viewConfigureProjectPanel configurationValues

        ProjectPanel project ->
            viewProjectPanel project



-- Supported panel types


defaultIconOpts : Icon.Options
defaultIconOpts =
    Icon.defaultOptions


{-| This is a panel that is used to for the the simple form to add a new project, its a bit of a CTA.
-}
viewAddProjectPanel : { value : String, dirty : Bool, problems : List String } -> Element Msg
viewAddProjectPanel repositoryField =
    viewPanelContainer
        [ viewNewProjectPanelHeader "New project" ( 2, 3 )
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
                            , scheme = Button.Secondary
                            , size = Button.Medium
                            , widthLength = fill
                            , heightLength = px 45
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
                            , heightLength = px 45
                            , disabled = not (List.isEmpty repositoryField.problems)
                            }
                        )
                    ]
                ]
            ]
        ]


viewAddKnownHostPanel : { gitUrl : GitUrl, publicKey : String } -> Element Msg
viewAddKnownHostPanel { gitUrl, publicKey } =
    viewPanelContainer
        [ viewNewProjectPanelHeader "New repository source" ( 3, 3 )
        , row
            [ width fill
            , spacingXY 0 20
            , paddingEach
                { top = 10
                , left = 5
                , right = 5
                , bottom = 0
                }
            ]
            [ el
                [ width <| fillPortion 2
                , height fill
                , Border.widthEach { left = 0, right = 1, top = 0, bottom = 0 }
                , Border.color Palette.neutral6
                , paddingEach { top = 0, left = 0, right = 10, bottom = 0 }
                ]
                (viewGitSourceLogo gitUrl { defaultIconOpts | size = 100, sizeUnit = Icon.Percentage })
            , column
                [ spacingXY 0 20
                , Font.color Palette.neutral3
                , Font.extraLight
                , width <| fillPortion 10
                , Font.size 15
                , Font.alignLeft
                , paddingEach { top = 10, left = 10, right = 0, bottom = 0 }
                ]
                [ paragraph []
                    [ text ("Because this is your first repository hosted on " ++ gitUrl.source ++ " we need you to ")
                    , text "tell Velocity their public SSH key."
                    ]
                , paragraph []
                    [ text "You can find their public SSH key by running"
                    ]
                , paragraph [ Font.color Palette.primary2, Font.family [ Font.monospace ] ]
                    [ text ("ssh-keyscan " ++ gitUrl.source ++ "") ]
                , paragraph []
                    [ text "If the results of the command match the sources published public key, "
                    , link [ Font.color Palette.primary1 ]
                        { url = "https://help.github.com/articles/github-s-ssh-key-fingerprints/"
                        , label = text "for example GitHub"
                        }
                    , text " then we can fully trust Velocity is talking to the correct servers."
                    ]
                ]
            ]
        , row [ width fill, height shrink ]
            [ el [ height (px 125), width fill, paddingEach { top = 20, left = 0, right = 0, bottom = 20 } ]
                (Input.multilineText
                    { leftIcon = Nothing
                    , rightIcon = Nothing
                    , label = Input.labelHidden "Public key"
                    , placeholder = Just <| "ssh-keyscan " ++ gitUrl.source
                    , dirty = False
                    , value = publicKey
                    , problems = []
                    , onChange = EnteredRepositorySourcePublicKey
                    }
                )
            ]
        , row [ width fill, spacingXY 10 0 ]
            [ el [ width fill ]
                (Button.button NewProjectBackButtonClicked
                    { rightIcon = Nothing
                    , centerRightIcon = Nothing
                    , leftIcon = Just Icon.arrowLeft
                    , centerLeftIcon = Nothing
                    , content = text "Back"
                    , scheme = Button.Secondary
                    , size = Button.Medium
                    , widthLength = fill
                    , heightLength = px 45
                    , disabled = False
                    }
                )
            , el [ width (fillPortion 2) ]
                (Button.button AddKnownHostButtonClicked
                    { rightIcon = Just Icon.arrowRight
                    , centerRightIcon = Nothing
                    , leftIcon = Nothing
                    , centerLeftIcon = Nothing
                    , content = text "Allow source"
                    , scheme = Button.Primary
                    , size = Button.Medium
                    , widthLength = fill
                    , heightLength = px 45
                    , disabled = False
                    }
                )
            ]
        ]


viewConfigureProjectPanel : { gitUrl : GitUrl, projectName : String } -> Element Msg
viewConfigureProjectPanel { gitUrl, projectName } =
    viewPanelContainer
        [ viewNewProjectPanelHeader "New project / configure" ( 2, 3 )
        , row
            [ width fill
            , spacingXY 10 20
            , Font.size 15
            , paddingEach
                { top = 10
                , left = 0
                , right = 0
                , bottom = 0
                }
            ]
            [ column [ width fill, spacingXY 0 10 ]
                [ wrappedRow [ height shrink, width fill, spacing 10 ]
                    [ paragraph [ width (fillPortion 2), Font.alignRight ] [ text "Name" ]
                    , paragraph
                        [ width (fillPortion 10)
                        , padding 10
                        , Font.alignLeft
                        , Background.color Palette.neutral7
                        , Border.rounded 5
                        , pointer
                        , Font.color Palette.primary4
                        , mouseOver
                            [ Background.color Palette.neutral6
                            , Font.color Palette.primary1
                            ]
                        , inFront
                            (el [ width shrink, alignRight, centerY, moveLeft 5 ] (Icon.edit2 Icon.defaultOptions))
                        ]
                        [ text gitUrl.fullName
                        ]
                    ]
                , wrappedRow [ height shrink, width fill, spacing 10 ]
                    [ paragraph [ width (fillPortion 2), Font.alignRight ] [ text "Source" ]
                    , paragraph
                        [ width (fillPortion 10)
                        , padding 10
                        , Font.alignLeft
                        , Font.color Palette.neutral3
                        , inFront
                            (el [ width shrink, alignRight, centerY, moveLeft 5 ]
                                (GitUrl.sourceIcon gitUrl { defaultIconOpts | size = 16 })
                            )
                        ]
                        [ text gitUrl.source ]
                    ]
                , wrappedRow [ height shrink, width fill, spacing 10 ]
                    [ paragraph [ width (fillPortion 2), Font.alignRight ] [ text "Repo" ]
                    , paragraph
                        [ width (fillPortion 10)
                        , padding 10
                        , Font.alignLeft
                        , Font.color Palette.neutral3
                        ]
                        [ text gitUrl.pathName ]
                    ]
                , wrappedRow [ height shrink, width fill, spacing 10 ]
                    [ paragraph [ width (fillPortion 2), Font.alignRight ] [ text "Public" ]
                    , paragraph
                        [ width (fillPortion 10)
                        , padding 10
                        , Font.alignLeft
                        , Font.color Palette.neutral3
                        ]
                        [ text "Yes" ]
                    ]
                ]
            ]
        , row [ width fill, spacingXY 10 0, paddingEach { top = 20, bottom = 0, left = 0, right = 0 } ]
            [ el [ width fill ]
                (Button.button NewProjectBackButtonClicked
                    { rightIcon = Nothing
                    , centerRightIcon = Nothing
                    , leftIcon = Just Icon.arrowLeft
                    , centerLeftIcon = Nothing
                    , content = text "Back"
                    , scheme = Button.Secondary
                    , size = Button.Medium
                    , widthLength = fill
                    , heightLength = px 45
                    , disabled = False
                    }
                )
            , el [ width (fillPortion 2) ]
                (Button.button CompleteProjectButtonClicked
                    { rightIcon = Just Icon.check
                    , centerRightIcon = Nothing
                    , leftIcon = Nothing
                    , centerLeftIcon = Nothing
                    , content = text "Complete project"
                    , scheme = Button.Primary
                    , size = Button.Medium
                    , widthLength = fill
                    , heightLength = px 45
                    , disabled = False
                    }
                )
            ]
        ]


viewGitSourceLogo : GitUrl -> Icon.Options -> Element msg
viewGitSourceLogo gitUrl opts =
    column [ height fill, width fill ]
        [ el
            [ centerX
            , centerY
            , Font.color Palette.primary2
            ]
            (GitUrl.sourceIcon gitUrl opts)
        , el
            [ centerX
            , centerY
            , padding 5
            , Background.color Palette.neutral6
            , Border.rounded 10
            , Font.size 12
            ]
            (text gitUrl.source)
        ]


viewNewProjectPanelHeader : String -> ( Int, Int ) -> Element msg
viewNewProjectPanelHeader headerText ( currentPanel, totalPanels ) =
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
        --        , inFront
        --            (Route.link
        --                [ width (px 20)
        --                , height (px 20)
        --                , Border.width 1
        --                , Border.rounded 5
        --                , Border.color Palette.neutral4
        --                , alignRight
        --                , mouseOver
        --                    [ Background.color Palette.neutral2
        --                    , Font.color Palette.white
        --                    ]
        --                ]
        --                (Icon.x Icon.defaultOptions)
        --                (Route.Home ActivePanel.None)
        --            )
        ]
        [ text headerText ]


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
                [ width (px 100)
                , height (px 100)
                ]
                [ Project.thumbnail project ]

        details =
            column
                [ width fill
                , height fill
                , Font.alignLeft
                ]
                [ row
                    [ width fill
                    , Font.extraLight
                    , Font.size 20
                    , Font.letterSpacing -0.5
                    , width fill
                    , Border.widthEach { bottom = 2, left = 0, top = 0, right = 0 }
                    , Border.color Palette.primary7
                    , paddingEach { top = 5, left = 0, bottom = 10, right = 0 }
                    , Font.color Palette.primary4
                    , spacingXY 10 0
                    ]
                    [ paragraph [ width fill ]
                        [ Route.link [ width fill, clip ] (text <| Project.name project) (Route.Project <| Project.slug project)
                        ]
                    , viewIf (Project.syncing project) <| column [ width shrink ] [ Loading.icon { width = 20, height = 20 } ]
                    ]
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
                        ]
                        [ newTabLink [ width fill ]
                            { url = Project.repository project
                            , label =
                                row []
                                    [ Icon.externalLink { defaultIconOpts | size = 15 }
                                    , el [ centerY, paddingEach { left = 5, right = 0, bottom = 0, top = 0 } ] (text "Open repo")
                                    ]
                            }
                        ]
                    , row [ width fill ] []
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
        , Border.color Palette.neutral6
        , Border.rounded 10
        , Background.color Palette.white
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
    | EnteredRepositorySourcePublicKey String
    | CompleteProjectButtonClicked
    | NewProjectBackButtonClicked
    | AddKnownHostButtonClicked
    | KnownHostCreated (Result Http.Error KnownHost)
    | ProjectCreated (Result Http.Error Project)
    | PassedSlowLoadThreshold
    | NoOp


update : Msg -> Model msg -> ( Model msg, Cmd Msg )
update msg model =
    let
        baseUrl =
            Context.baseUrl model.context
    in
    case msg of
        NoOp ->
            ( model, Cmd.none )

        NewProjectBackButtonClicked ->
            ( { model
                | projectFormStatus =
                    case model.projectFormStatus of
                        NotOpen ->
                            NotOpen

                        SettingRepository _ ->
                            NotOpen

                        AddingKnownHost { gitUrl } ->
                            SettingRepository { value = gitUrl.href, dirty = True, problems = [] }

                        ConfiguringRepository { gitUrl } ->
                            SettingRepository { value = gitUrl.href, dirty = True, problems = [] }
              }
            , Cmd.none
            )

        AddKnownHostButtonClicked ->
            ( model
            , case ( model.projectFormStatus, Session.cred model.session ) of
                ( AddingKnownHost { publicKey }, Just cred ) ->
                    KnownHost.create cred baseUrl publicKey
                        |> Http.send KnownHostCreated

                _ ->
                    Cmd.none
            )

        CompleteProjectButtonClicked ->
            ( model
            , case ( model.projectFormStatus, Session.cred model.session ) of
                ( ConfiguringRepository { gitUrl, projectName }, Just cred ) ->
                    Project.create cred
                        baseUrl
                        { name = gitUrl.fullName
                        , repository = gitUrl.href
                        , privateKey = Nothing
                        }
                        |> Http.send ProjectCreated

                _ ->
                    Cmd.none
            )

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

        EnteredRepositorySourcePublicKey publicKey ->
            ( { model
                | projectFormStatus =
                    case model.projectFormStatus of
                        AddingKnownHost internals ->
                            AddingKnownHost { internals | publicKey = publicKey }

                        _ ->
                            model.projectFormStatus
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

        ParsedRepository (Ok { gitUrl }) ->
            ( { model
                | projectFormStatus =
                    if KnownHost.isUnknownHost (Session.knownHosts model.session) (Just gitUrl) then
                        AddingKnownHost
                            { gitUrl = gitUrl
                            , publicKey = ""
                            }

                    else
                        ConfiguringRepository
                            { gitUrl = gitUrl
                            , projectName = gitUrl.fullName
                            }
              }
            , Cmd.none
            )

        ProjectCreated (Ok project) ->
            ( { model
                | session = Session.addProject project model.session
                , projectFormStatus = NotOpen
              }
            , case Session.cred model.session of
                Just cred ->
                    Project.sync cred (Context.baseUrl model.context) (Project.slug project)
                        |> Http.send (always NoOp)

                Nothing ->
                    Cmd.none
            )

        KnownHostCreated (Ok knownHost) ->
            ( { model
                | session = Session.addKnownHost knownHost model.session
                , projectFormStatus =
                    case model.projectFormStatus of
                        AddingKnownHost { gitUrl } ->
                            ConfiguringRepository
                                { gitUrl = gitUrl
                                , projectName = gitUrl.fullName
                                }

                        _ ->
                            model.projectFormStatus
              }
            , Cmd.none
            )

        ProjectCreated (Err _) ->
            ( model, Cmd.none )

        KnownHostCreated (Err _) ->
            ( model, Cmd.none )

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


subscriptions : Model msg -> Sub Msg
subscriptions model =
    Sub.batch
        [ Session.changes UpdateSession model.context model.session
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


toSession : Model msg -> Session
toSession model =
    model.session


toContext : Model msg -> Context msg
toContext model =
    model.context



-- UTIL


viewIf : Bool -> Element msg -> Element msg
viewIf condition content =
    if condition then
        content

    else
        none
