module Page.Project exposing (Model, Msg, init, subscriptions, toContext, toSession, update, view)

import Asset
import Browser.Events
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Form.Input as Input
import Form.Select as Select
import Icon
import Json.Decode as Decode
import Palette
import Project exposing (Project)
import Project.Id exposing (Id)
import Session exposing (Session)



-- Model


type alias Model msg =
    { session : Session
    , context : Context msg
    , id : Id
    , branchDropdown : BranchDropdown
    }


type BranchDropdown
    = BranchDropdown BranchDropdownStatus


type BranchDropdownStatus
    = Open
    | ListenClicks
    | Closed


init : Session -> Context msg -> Id -> ( Model msg, Cmd Msg )
init session context projectId =
    ( { session = session
      , context = context
      , id = projectId
      , branchDropdown = BranchDropdown Closed
      }
    , Cmd.none
    )



-- Subscriptions


subscriptions : Model msg -> Sub Msg
subscriptions { branchDropdown } =
    case branchDropdown of
        BranchDropdown Open ->
            Browser.Events.onAnimationFrame
                (\_ -> BranchDropdownListenClicks)

        BranchDropdown ListenClicks ->
            Browser.Events.onClick
                (Decode.succeed BranchDropdownClose)

        BranchDropdown Closed ->
            Sub.none



-- Update


type Msg
    = NoOp
    | BranchDropdownToggleClicked
    | BranchDropdownListenClicks
    | BranchDropdownClose


update : Msg -> Model msg -> ( Model msg, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        BranchDropdownToggleClicked ->
            let
                state =
                    if model.branchDropdown == BranchDropdown Open then
                        BranchDropdown Closed

                    else
                        BranchDropdown Open
            in
            ( { model | branchDropdown = state }
            , Cmd.none
            )

        BranchDropdownListenClicks ->
            ( { model | branchDropdown = BranchDropdown ListenClicks }
            , Cmd.none
            )

        BranchDropdownClose ->
            ( { model | branchDropdown = BranchDropdown Closed }
            , Cmd.none
            )



-- EXPORT


toSession : Model msg -> Session
toSession model =
    model.session


toContext : Model msg -> Context msg
toContext model =
    model.context



-- View


view : Model msg -> { title : String, content : Element Msg }
view model =
    let
        device =
            Context.device model.context

        projects =
            Session.projects model.session

        maybeProject =
            Project.findProject projects model.id
    in
    { title = "Project page"
    , content =
        case maybeProject of
            Just project ->
                column [ width fill, height fill ]
                    [ viewSubHeader device project model
                    , viewBody device model project
                    ]

            Nothing ->
                viewErrored
    }


viewErrored : Element msg
viewErrored =
    text "Something went seriously wrong"



-- SubHeader


viewSubHeader : Device -> Project -> Model msg -> Element Msg
viewSubHeader device project model =
    case device.class of
        Phone ->
            viewMobileSubHeader project model

        Tablet ->
            viewDesktopSubHeader project model

        Desktop ->
            viewDesktopSubHeader project model

        BigDesktop ->
            viewDesktopSubHeader project model


viewMobileSubHeader : Project -> Model msg -> Element Msg
viewMobileSubHeader project model =
    row
        [ width fill
        , height shrink
        , Background.color Palette.white
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
        , below
            (if model.branchDropdown == BranchDropdown ListenClicks then
                el
                    [ width fill
                    , height (px 9999)
                    , moveDown 1
                    , Background.color Palette.neutral7
                    , Font.color Palette.primary5
                    ]
                    viewBranchSelectDropdown

             else
                none
            )
        ]
        [ row
            [ width fill
            , centerY
            , Font.color Palette.neutral3
            , spaceEvenly
            ]
            [ el [ width fill, Font.alignLeft, paddingXY 20 0 ] (text <| Project.name project)
            , el [ width fill, paddingXY 20 0 ] (viewBranchSelectButton fill model.branchDropdown)
            ]
        ]


viewDesktopSubHeader : Project -> Model msg -> Element Msg
viewDesktopSubHeader project model =
    row
        [ Font.bold
        , Font.size 18
        , width (fill |> maximum 1600)
        , alignRight
        , height (px 65)
        , paddingXY 20 10
        , Font.color Palette.white
        , Background.color Palette.white
        , Font.color Palette.white
        , Border.widthEach { top = 1, bottom = 1, left = 0, right = 0 }
        , Border.color Palette.neutral6
        ]
        [ row
            [ width fill
            , centerY
            , Font.color Palette.neutral3
            ]
            [ el [ width fill ] <|
                el [ alignLeft ] (text <| Project.name project)
            , el
                [ width fill
                , below
                    (if model.branchDropdown == BranchDropdown ListenClicks then
                        column
                            [ width (fill |> maximum 600 |> minimum 400)
                            , alignRight
                            , Font.size 14
                            , height shrink
                            , Background.color Palette.neutral7
                            , Border.width 1
                            , Border.color Palette.neutral6
                            , Border.rounded 5
                            , moveDown 3
                            , Border.shadow
                                { offset = ( 2, 2 )
                                , size = 1
                                , blur = 2
                                , color = Palette.neutral4
                                }
                            ]
                            [ column
                                [ Border.widthEach { top = 0, left = 0, right = 0, bottom = 1 }
                                , Border.color Palette.neutral6
                                , width fill
                                ]
                                [ row
                                    [ width fill
                                    , spaceEvenly
                                    , paddingEach { bottom = 0, left = 10, right = 10, top = 10 }
                                    ]
                                    [ el
                                        [ alignLeft
                                        , centerY
                                        ]
                                        (text "Switch branches")
                                    , el
                                        [ alignRight
                                        , centerY
                                        , Font.color Palette.neutral4
                                        , pointer
                                        , mouseOver [ Font.color Palette.neutral1 ]
                                        ]
                                        (Icon.x Icon.defaultOptions)
                                    ]
                                , viewBranchSelectDropdown
                                ]
                            ]

                     else
                        none
                    )
                ]
                (el
                    [ alignRight
                    ]
                    (viewBranchSelectButton shrink model.branchDropdown)
                )
            ]
        ]



-- Body


viewBody : Device -> Model msg -> Project -> Element Msg
viewBody device model project =
    case device.class of
        Phone ->
            viewMobileBody model project

        Tablet ->
            viewDesktopBody model project

        Desktop ->
            viewDesktopBody model project

        BigDesktop ->
            viewBigDesktopBody model project


viewMobileBody : Model msg -> Project -> Element Msg
viewMobileBody model project =
    column
        [ width fill
        , height fill
        , Background.color Palette.neutral7
        , centerX
        , spacing 20
        , padding 20
        ]
        [ viewProjectDetails project
        , viewProjectHealth project
        , viewTimeline
        ]


viewDesktopBody : Model msg -> Project -> Element Msg
viewDesktopBody model project =
    row
        [ height fill
        , Background.color Palette.neutral7
        , centerX
        , spacing 20
        , padding 20
        , width (fill |> maximum 1600)
        , alignRight
        ]
        [ column
            [ width (fillPortion 4)
            , height fill
            , spacingXY 0 20
            ]
            [ viewProjectDetails project
            , viewProjectHealth project
            ]
        , column
            [ width (fillPortion 6)
            , height fill
            ]
            [ viewTimeline ]
        ]


viewBigDesktopBody : Model msg -> Project -> Element Msg
viewBigDesktopBody model project =
    row
        [ height fill
        , Background.color Palette.neutral7
        , centerX
        , spacing 20
        , padding 20
        , width (fill |> maximum 1600)
        , alignRight
        ]
        [ column
            [ width (fillPortion 3)
            , height fill
            ]
            [ viewProjectDetails project
            ]
        , column
            [ width (fillPortion 3)
            , height fill
            ]
            [ viewProjectHealth project
            ]
        , column
            [ width (fillPortion 6)
            , height fill
            ]
            [ viewTimeline ]
        ]



-- Project Details


viewProjectDetails : Project -> Element Msg
viewProjectDetails project =
    column
        [ width fill
        , height shrink
        , Background.color Palette.white
        , Border.shadow
            { offset = ( 1, 1 )
            , size = 1
            , blur = 1
            , color = Palette.neutral6
            }
        , Border.rounded 5
        ]
        [ el
            [ Font.size 20
            , width fill
            , Font.alignLeft
            , paddingXY 10 20
            ]
            (text "Details")
        , row
            [ width fill
            , Border.widthEach { top = 1, bottom = 0, left = 0, right = 0 }
            , paddingXY 10 0
            , height (px 50)
            , Border.color Palette.neutral6
            , Font.size 15
            , spaceEvenly
            ]
            [ el [ Font.color Palette.neutral2 ] (text "Repository")
            , row
                [ spacingXY 5 0
                , Font.color Palette.primary3
                , pointer
                , mouseOver [ Font.color Palette.primary4 ]
                ]
                [ el [] (Icon.github Icon.defaultOptions)
                , el [] (text (Project.name project))
                ]
            ]
        , row
            [ width fill
            , Border.widthEach { top = 1, bottom = 0, left = 0, right = 0 }
            , paddingXY 10 0
            , Border.color Palette.neutral6
            , Font.size 15
            , spaceEvenly
            , height (px 50)
            , inFront
                (el
                    [ padding 5
                    , Border.rounded 5
                    , Background.color Palette.neutral6
                    , alignRight
                    , centerY
                    , moveLeft 10
                    , Font.size 13
                    , Font.family [ Font.monospace ]
                    ]
                    (text "master")
                )
            ]
            [ el [ Font.color Palette.neutral2 ] (text "Default branch")
            ]
        ]



-- Project Health


viewProjectHealth : Project -> Element Msg
viewProjectHealth project =
    column
        [ width fill
        , height shrink
        , Background.color Palette.white
        , Border.shadow
            { offset = ( 1, 1 )
            , size = 1
            , blur = 1
            , color = Palette.neutral6
            }
        , Border.rounded 5
        ]
        [ el
            [ Font.size 20
            , width fill
            , Font.alignLeft
            , paddingXY 10 20
            ]
            (text "Health")
        , row
            [ width fill
            , Border.widthEach { top = 1, bottom = 0, left = 0, right = 0 }
            , paddingXY 10 0
            , height (px 50)
            , Border.color Palette.neutral6
            , Font.size 15
            , spaceEvenly
            ]
            [ el [ Font.color Palette.neutral2 ] (text "run-unit-tests")
            , el [ Font.color Palette.success3 ] (Icon.sun Icon.defaultOptions)
            ]
        , row
            [ width fill
            , Border.widthEach { top = 1, bottom = 0, left = 0, right = 0 }
            , paddingXY 10 0
            , height (px 50)
            , Border.color Palette.neutral6
            , Font.size 15
            , spaceEvenly
            ]
            [ el [ Font.color Palette.neutral2 ] (text "deploy-master")
            , el [ Font.color Palette.warning3 ] (Icon.cloudRain Icon.defaultOptions)
            ]
        ]



-- Task Container


viewTasksContainer : Model msg -> Element Msg
viewTasksContainer model =
    column
        [ width fill
        , height shrink
        , padding 10
        ]
        [ viewTasksList
        ]


viewBranchSelectButton : Length -> BranchDropdown -> Element Msg
viewBranchSelectButton widthLength (BranchDropdown state) =
    el
        [ width fill
        , height fill
        ]
        (Button.button BranchDropdownToggleClicked
            { leftIcon = Nothing
            , rightIcon = Nothing
            , centerLeftIcon = Nothing
            , centerRightIcon = Nothing
            , scheme =
                if state == Closed then
                    Button.Secondary

                else
                    Button.Primary
            , content =
                row [ spacingXY 5 0 ]
                    [ el [] (text "Branch:")
                    , el [ Font.heavy ] (text "master")
                    , el [] (Icon.arrowDown { size = 12, sizeUnit = Icon.Pixels, strokeWidth = 1 })
                    ]
            , size = Button.Small
            , widthLength = widthLength
            , heightLength = shrink
            , disabled = False
            }
        )


viewBranchSelectDropdown : Element Msg
viewBranchSelectDropdown =
    column [ width fill ]
        [ row
            [ width fill
            , padding 10
            ]
            [ el
                [ Background.color Palette.white
                , width fill
                ]
                (Input.search
                    { leftIcon = Just Icon.search
                    , rightIcon = Nothing
                    , label = Input.labelHidden "Search for a branch"
                    , placeholder = Just "Find a branch..."
                    , dirty = False
                    , value = ""
                    , problems = []
                    , onChange = always NoOp
                    }
                )
            ]
        , column
            [ width fill
            , paddingEach { top = 0, left = 0, right = 0, bottom = 0 }
            ]
            [ row
                [ Border.widthEach { top = 1, bottom = 0, right = 0, left = 0 }
                , Border.color Palette.neutral6
                , width fill
                , padding 10
                , Background.color Palette.white
                , spacingXY 10 0
                , Font.color Palette.neutral1
                , pointer
                , mouseOver
                    [ Background.color Palette.primary4
                    , Font.color Palette.neutral7
                    ]
                ]
                [ el [ width shrink, centerY ] (Icon.check Icon.defaultOptions)
                , el [ width fill, centerY, Font.alignLeft, clipX ] (text "master")
                ]
            , row
                [ Border.widthEach { top = 1, bottom = 0, right = 0, left = 0 }
                , Border.color Palette.neutral6
                , width fill
                , padding 10
                , Background.color Palette.white
                , spacingXY 10 0
                , Font.color Palette.neutral1
                , Border.roundEach { bottomLeft = 5, bottomRight = 5, topLeft = 0, topRight = 0 }
                , pointer
                , mouseOver
                    [ Background.color Palette.primary4
                    , Font.color Palette.neutral7
                    ]
                ]
                [ el [ width (px 16), centerY ] none
                , el [ width fill, centerY, Font.alignLeft, clipX ] (text "elm-upgrade")
                ]
            ]
        ]


viewTasksList : Element msg
viewTasksList =
    column [ width fill, spacingXY 0 10, paddingXY 0 10 ]
        []


viewTask : Element msg
viewTask =
    row
        [ width fill
        , padding 15
        , Background.color Palette.neutral7
        ]
        [ text "run-unit-tests" ]



-- Timeline


avatarIcon : Element msg
avatarIcon =
    el
        [ width (px 30)
        , height (px 30)
        , Border.rounded 180
        , Background.image <| Asset.src Asset.defaultAvatar
        ]
        none


viewTimeline : Element msg
viewTimeline =
    column
        [ width fill
        , height shrink
        , Background.color Palette.white
        , Border.shadow
            { offset = ( 1, 1 )
            , size = 1
            , blur = 1
            , color = Palette.neutral6
            }
        , Border.rounded 5
        , Font.size 14
        , Font.alignLeft
        ]
        [ el
            [ Font.size 20
            , width fill
            , Font.alignLeft
            , paddingXY 10 20
            , Border.widthEach { bottom = 1, right = 0, top = 0, left = 0 }
            , Border.color Palette.neutral6
            ]
            (text "Timeline")
        , column
            [ width fill
            , behindContent
                (el
                    [ height fill
                    , width (px 1)
                    , Background.color Palette.neutral6
                    , moveRight 24
                    ]
                    none
                )
            ]
            [ viewCommitListRowOne { top = 0, bottom = 1 }
            , viewCommitListRowTwo { top = 0, bottom = 1 }
            , viewCommitListRowThree { top = 0, bottom = 0 }
            ]
        ]


viewTimelineRow : { topBorder : Int, bottomBorder : Int } -> Element msg -> Element msg
viewTimelineRow { topBorder, bottomBorder } content =
    content


viewCommitListRowOne : { top : Int, bottom : Int } -> Element msg
viewCommitListRowOne { top, bottom } =
    row
        [ spacingXY 15 0
        , width fill
        , Border.widthEach { bottom = bottom, top = top, left = 0, right = 0 }
        , paddingXY 10 17
        , Border.color Palette.neutral6
        , height (px 50)
        ]
        [ el
            [ width shrink
            , height shrink
            ]
            avatarIcon
        , wrappedRow [ width fill, height shrink, spacing 5 ]
            [ el [ Font.color Palette.neutral1, Font.heavy ] (text "VJ")
            , el [ Font.color Palette.neutral3 ] (text "pushed")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "WIP - Frontend compiling")
            , el [ Font.color Palette.neutral3 ] (text "to")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "CP-179-multicargo")
            , el [ Font.size 14, Font.color Palette.neutral5, alignRight ] (text "2 days ago")
            ]
        ]


viewCommitListRowTwo : { top : Int, bottom : Int } -> Element msg
viewCommitListRowTwo { top, bottom } =
    row
        [ spacingXY 15 0
        , width fill
        , Border.widthEach { bottom = bottom, top = top, left = 0, right = 0 }
        , paddingXY 10 17
        , Border.color Palette.neutral6
        , height (px 50)
        ]
        [ el
            [ width shrink
            , height shrink
            ]
            avatarIcon
        , wrappedRow [ width fill, height shrink, spacing 5 ]
            [ el
                [ padding 5
                , Background.color Palette.success5
                , Font.color Palette.white
                , Font.heavy
                , Font.size 16
                , Font.variant Font.smallCaps
                , Border.rounded 5
                , width (px 60)
                ]
                (el
                    [ centerX
                    , centerY
                    ]
                    (text "success")
                )
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "Eddy")
            , el [ Font.color Palette.neutral3 ] (text "ran")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "run-unit-tests")
            , el [ Font.color Palette.neutral3 ] (text "on")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "master")
            , el [ Font.size 14, Font.color Palette.neutral5, alignRight ] (text "2 days ago")
            ]
        ]


viewCommitListRowThree : { top : Int, bottom : Int } -> Element msg
viewCommitListRowThree { top, bottom } =
    row
        [ spacingXY 15 0
        , width fill
        , Border.widthEach { bottom = bottom, top = top, left = 0, right = 0 }
        , paddingXY 10 17
        , Border.color Palette.neutral6
        , height (px 50)
        ]
        [ el
            [ width shrink
            , height shrink
            ]
            avatarIcon
        , wrappedRow [ width fill, height shrink, spacing 5 ]
            [ el
                [ padding 5
                , Background.color Palette.danger5
                , Font.color Palette.white
                , Font.heavy
                , Font.size 16
                , Font.variant Font.smallCaps
                , Border.rounded 5
                , width (px 60)
                ]
                (el
                    [ centerX
                    , centerY
                    ]
                    (text "failure")
                )
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "Eddy")
            , el [ Font.color Palette.neutral3 ] (text "ran")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "run-unit-tests")
            , el [ Font.color Palette.neutral3 ] (text "on")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "master")
            , el [ Font.size 14, Font.color Palette.neutral5, alignRight ] (text "2 days ago")
            ]
        ]
