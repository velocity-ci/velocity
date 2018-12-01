module Page.Project exposing (Model, Msg, init, toContext, toSession, update, view)

import Asset
import Context exposing (Context)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Form.Input as Input
import Icon
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
    = BranchDropdown BranchDropdownState


type BranchDropdownState
    = Open
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



-- Update


type Msg
    = NoOp
    | BranchDropdownToggleClicked


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
                    [ viewSubHeader device project
                    , row
                        [ width fill
                        , height fill
                        , Background.color Palette.white
                        , centerX
                        , spacing 20
                        , padding 20
                        ]
                        [ column [ width (fillPortion 4), height fill ]
                            [ viewTasksContainer model ]
                        , column [ width (fillPortion 6), height fill ]
                            [ viewCommitsList ]
                        ]
                    ]

            Nothing ->
                viewErrored
    }


viewErrored : Element msg
viewErrored =
    text "Something went seriously wrong"



-- SubHeader


viewSubHeader : Device -> Project -> Element msg
viewSubHeader device project =
    case device.class of
        Phone ->
            viewMobileSubHeader project

        Tablet ->
            viewDesktopSubHeader project

        Desktop ->
            viewDesktopSubHeader project

        BigDesktop ->
            viewDesktopSubHeader project


viewMobileSubHeader : Project -> Element msg
viewMobileSubHeader project =
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
        , el [ width fill ] none
        , el [ width fill ] none
        ]


viewDesktopSubHeader : Project -> Element msg
viewDesktopSubHeader project =
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
            (el [ alignLeft ] (text <| Project.name project))
        ]



-- Task Container


viewTasksContainer : Model msg -> Element Msg
viewTasksContainer model =
    column
        [ width fill
        , height shrink
        , padding 10
        ]
        [ viewBranchSelectButton model.branchDropdown
        ]


viewBranchSelectButton : BranchDropdown -> Element Msg
viewBranchSelectButton (BranchDropdown state) =
    el
        [ width fill
        , below
            (if state == Open then
                viewBranchSelectDropdown

             else
                none
            )
        ]
        (el
            [ alignRight
            ]
            (Button.button BranchDropdownToggleClicked
                { leftIcon = Nothing
                , rightIcon = Nothing
                , centerLeftIcon = Nothing
                , centerRightIcon = Nothing
                , scheme = Button.Secondary
                , content =
                    row [ spacingXY 5 0 ]
                        [ el [ Font.color Palette.neutral3 ] (text "Branch:")
                        , el [ Font.semiBold, Font.color Palette.neutral2 ] (text "master")
                        , el [ Font.color Palette.neutral2 ] (Icon.arrowDown { size = 12, sizeUnit = Icon.Pixels, strokeWidth = 1 })
                        ]
                , size = Button.Small
                , widthLength = shrink
                , heightLength = shrink
                , disabled = False
                }
            )
        )


viewBranchSelectDropdown : Element Msg
viewBranchSelectDropdown =
    row [ width fill ]
        [ column
            [ width (fillPortion 9)
            , height shrink
            , Background.color Palette.neutral7
            , Border.width 1
            , Border.color Palette.neutral6
            , Border.rounded 5
            , moveDown 3
            , Font.size 14
            ]
            [ row
                [ Border.widthEach { top = 0, left = 0, right = 0, bottom = 1 }
                , Border.color Palette.neutral6
                , width fill
                , padding 10
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
            , row
                [ width fill
                , padding 10
                ]
                [ el
                    [ Background.color Palette.white
                    , width fill
                    , height (px 25)
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
                , paddingEach { top = 15, left = 0, right = 0, bottom = 0 }
                ]
                [ row
                    [ Border.widthEach { top = 1, bottom = 0, right = 0, left = 0 }
                    , Border.color Palette.neutral6
                    , width fill
                    , padding 10
                    , Background.color Palette.white
                    , spacingXY 10 0
                    , Font.color Palette.neutral1
                    , Border.rounded 5
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
                    , Border.rounded 5
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
        ]


viewTask : Element msg
viewTask =
    row
        [ width fill
        , height (px 200)
        , Background.color Palette.neutral7
        ]
        []



-- Commits


avatarIcon : Element msg
avatarIcon =
    el
        [ width (px 30)
        , height (px 30)
        , Border.rounded 180
        , Background.image <| Asset.src Asset.defaultAvatar
        ]
        none


viewCommitsList : Element msg
viewCommitsList =
    column
        [ Border.width 0
        , Border.color Palette.neutral6
        , Border.rounded 10
        , width fill
        , height shrink
        , padding 10
        , Font.size 14
        , Font.alignLeft
        ]
        [ el
            [ width fill
            , Font.size 18
            , paddingEach { top = 0, left = 0, right = 0, bottom = 20 }
            , Border.widthEach { top = 0, left = 0, right = 0, bottom = 1 }
            , Border.color Palette.neutral6
            , Font.heavy
            , Font.color Palette.primary1
            ]
            (text "Timeline")
        , viewCommitListRow
        , viewCommitListRow
        , viewCommitListRow
        ]


viewCommitListRow : Element msg
viewCommitListRow =
    row
        [ spacingXY 10 0
        , width fill
        , Border.widthEach { bottom = 1, top = 0, left = 0, right = 0 }
        , paddingXY 0 17
        , Border.color Palette.neutral6
        ]
        [ el
            [ width shrink
            , height shrink
            ]
            avatarIcon
        , wrappedRow [ width fill, height shrink, spacing 5 ]
            [ el [ Font.color Palette.neutral1, Font.heavy ] (text "Edd L")
            , el [ Font.color Palette.neutral3 ] (text "pushed")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "WIP - Frontend compiling")
            , el [ Font.color Palette.neutral3 ] (text "to")
            , el [ Font.color Palette.neutral1, Font.heavy ] (text "CP-179-multicargo")
            ]
        ]
