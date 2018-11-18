port module Element.ProjectForm exposing (State, init, isConfiguring, parseGitUrlCmd, subscriptions, view)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Form
import Form.Input
import GitUrl exposing (GitUrl)
import Html.Attributes
import Icon
import Json.Decode as Decode
import Json.Encode as Encode
import Page.Home.ActivePanel as ActivePanel
import Palette
import Route



-- PORTS


port parseGitUrl : ( String, Bool ) -> Cmd msg


port onGitUrlParsed : (Encode.Value -> msg) -> Sub msg



-- TYPES


type State
    = CheckingUrl Url
    | CheckedUrl Url GitUrl
    | ProjectConfigurartion Url GitUrl


type Url
    = Url String


type Name
    = Name String


init : State
init =
    CheckingUrl (Url "")



-- INFO


isConfiguring : State -> Bool
isConfiguring state =
    case state of
        ProjectConfigurartion _ _ ->
            True

        _ ->
            False



-- SUBSCRIPTIONS


type alias GitUrlResponse =
    { gitUrl : String
    , parsed : Maybe GitUrl
    , configuring : Bool
    }


subscriptions : (State -> Cmd msg -> msg) -> Sub msg
subscriptions updateMsg =
    onGitUrlParsed
        (\encoded ->
            let
                decoder =
                    Decode.map3 GitUrlResponse
                        (Decode.field "gitUrl" Decode.string)
                        (Decode.maybe (Decode.field "parsed" GitUrl.decoder))
                        (Decode.field "configuring" Decode.bool)

                noOp =
                    updateMsg (CheckingUrl (Url "")) Cmd.none
            in
            case Decode.decodeValue decoder encoded of
                Ok { gitUrl, parsed, configuring } ->
                    let
                        _ =
                            Debug.log "gitUrl" gitUrl

                        _ =
                            Debug.log "parsed" parsed
                    in
                    case ( parsed, configuring ) of
                        ( Just parsedGirUrl, True ) ->
                            updateMsg (ProjectConfigurartion (Url gitUrl) parsedGirUrl) Cmd.none

                        ( Just parsedGitUrl, False ) ->
                            if String.isEmpty parsedGitUrl.resource then
                                updateMsg (CheckingUrl (Url gitUrl)) Cmd.none

                            else
                                updateMsg (CheckedUrl (Url gitUrl) parsedGitUrl) Cmd.none

                        _ ->
                            updateMsg (CheckingUrl (Url gitUrl)) Cmd.none

                Err _ ->
                    noOp
        )



--parsedHandler (Cmd msg -> msg) -> Sub msg
-- VIEW


defaultIconOpts : Icon.Options
defaultIconOpts =
    Icon.defaultOptions


view : State -> (State -> Cmd msg -> msg) -> Element msg
view state updateMsg =
    case state of
        CheckingUrl url ->
            viewFirstPanel url Nothing state updateMsg

        CheckedUrl url gitUrl ->
            viewFirstPanel url (Just gitUrl) state updateMsg

        ProjectConfigurartion url gitUrl ->
            viewSecondPanel url gitUrl state updateMsg


viewFirstPanel : Url -> Maybe GitUrl -> State -> (State -> Cmd msg -> msg) -> Element msg
viewFirstPanel url maybeGitUrl state updateMsg =
    column [ spacingXY 0 20, width fill ]
        [ viewHelpText
        , column [ width fill ]
            [--             viewUrlField url maybeGitUrl (\newUrl parseCmd -> updateMsg (CheckingUrl newUrl) parseCmd)
            ]
        , row [ width fill ]
            [ el [ width fill ] none
            , el [ width (fillPortion 3) ] (viewNextButton state updateMsg)
            ]
        ]


viewSummaryPanel : GitUrl -> Element msg
viewSummaryPanel gitUrl =
    column [ spacing 10, width fill ]
        [ viewSummaryRow "Home" gitUrl.fullName
        , el
            [ Font.size 16
            , Font.color Palette.primary3
            , width fill
            , height shrink
            , alignRight
            , inFront
                (newTabLink []
                    { url = gitUrl.href
                    , label = Icon.externalLink Icon.defaultOptions
                    }
                )
            ]
            none
        , wrappedRow [ spacing 10, width fill ]
            [ column [ spacingXY 0 5, width shrink ]
                [ column [ width shrink ] [ el [ Font.size 18, Font.color Palette.neutral5 ] (text "Source") ]
                , column [ width shrink ] [ el [ Font.size 16, Font.color Palette.primary5 ] (Icon.github Icon.defaultOptions) ]
                ]
            , column [ spacingXY 0 5, width fill ]
                [ column [ width fill ] [ el [ Font.size 18, Font.color Palette.neutral5 ] (text "Protocol") ]
                , column [ width fill ] [ el [ Font.size 16, Font.color Palette.primary5 ] (text gitUrl.protocol) ]
                ]
            ]
        ]


viewSummaryRow : String -> String -> Element msg
viewSummaryRow label value =
    column [ spacingXY 0 5, width fill ]
        [ column [ width fill ]
            [ el [ Font.size 18, Font.color Palette.neutral5 ]
                (text label)
            ]
        , column [ width fill ]
            [ el [ Font.size 16, Font.color Palette.primary3 ]
                (text value)
            ]
        ]


viewHostIcon : String -> Element msg
viewHostIcon resource =
    case resource of
        "github.com" ->
            Icon.github Icon.defaultOptions

        "gitlab.com" ->
            Icon.gitlab Icon.defaultOptions

        _ ->
            Icon.gitPullRequest Icon.defaultOptions


viewSecondPanel : Url -> GitUrl -> State -> (State -> Cmd msg -> msg) -> Element msg
viewSecondPanel url gitUrl state updateMsg =
    column [ spacingXY 0 20, width fill ]
        [ column [ width fill ] [ viewSummaryPanel gitUrl ]
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


viewUrlField : Url -> Maybe GitUrl -> (Url -> Cmd msg -> msg) -> Element msg
viewUrlField (Url val) maybeGitUrl updateMsg =
    let
        isValid =
            Maybe.map (always True) maybeGitUrl
                |> Maybe.withDefault False

        isDirty =
            not <| String.isEmpty val

        statusColourAttrs =
            Form.statusAttrs isValid isDirty

        leftIcon =
            el
                [ width shrink
                , height shrink
                , alignLeft
                , moveRight 7
                , centerY
                ]
            <|
                Icon.link { defaultIconOpts | size = 16 }

        rightIcon =
            if maybeGitUrl /= Nothing then
                el
                    [ width shrink
                    , height shrink
                    , alignRight
                    , moveLeft 7
                    , centerY
                    ]
                <|
                    Icon.check { defaultIconOpts | size = 16 }

            else
                none

        placeholder =
            Input.placeholder
                [ width shrink
                , height shrink
                , centerY
                , Font.color Palette.neutral4
                ]
                (text "Repository URL")
    in
    row
        (statusColourAttrs
            ++ [ height (px 40)
               , width fill
               , Border.width 1
               , Border.rounded 4
               , mouseOver [ Border.shadow { offset = ( 0, 0 ), size = 1, blur = 0, color = Palette.neutral5 } ]
               , Font.size 16
               , inFront leftIcon
               , inFront rightIcon
               ]
        )
        [ Input.text
            [ Input.focusedOnLoad
            , Border.width 0
            , Background.color Palette.transparent
            , paddingXY 30 0
            , height fill
            , focused (Form.statusDecorations isValid isDirty)
            ]
            { onChange = \value -> updateMsg (Url val) (parseGitUrlCmd value False)
            , placeholder = Just placeholder
            , text = val
            , label = Input.labelHidden "Repository URL"
            }
        ]


parseGitUrlCmd : String -> Bool -> Cmd msg
parseGitUrlCmd value configuring =
    parseGitUrl ( value, configuring )


viewNextButton : State -> (State -> Cmd msg -> msg) -> Element msg
viewNextButton state updateState =
    let
        iconOpts =
            { defaultIconOpts | sizeUnit = Icon.Pixels, size = 22.5 }
    in
    case state of
        CheckedUrl (Url val) _ ->
            Route.link
                Form.buttonAttrs
                (row
                    [ width fill
                    , inFront
                        (el
                            [ alignRight
                            , width shrink
                            , height shrink
                            , moveUp 4
                            , alignTop
                            ]
                            (Icon.arrowRight iconOpts)
                        )
                    ]
                    [ row [ centerX ]
                        [ Icon.settings { defaultIconOpts | size = 15 }
                        , el [ paddingXY 5 0, Font.heavy ] (text "Configure")
                        ]
                    ]
                )
                (Route.Home
                    (Just <| ActivePanel.ConfigureProjectForm val)
                )

        _ ->
            none
