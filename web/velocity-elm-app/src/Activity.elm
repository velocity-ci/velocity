module Activity exposing (Log, init, projectAdded, view)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Palette
import Project exposing (Project)
import Project.Id as Project



-- TYPES


type Activity
    = Project CategoryProject


type CategoryProject
    = ProjectAdded Project.Id



-- LOG


type Log
    = Log (List Activity)


init : Log
init =
    Log []


projectAdded : Project.Id -> Log -> Log
projectAdded id (Log log) =
    Log <| Project (ProjectAdded id) :: log



---- VIEW


type alias Entities =
    { projects : List Project }


view : Entities -> Log -> Element msg
view entities (Log activities) =
    column
        [ Background.color Palette.primary2
        , width fill
        , height fill
        , paddingEach { top = 80, bottom = 90, left = 20, right = 20 }
        , spacing 10
        ]
        [ viewNotificationsPanelHeading
        , viewNotificationCommitItem
        , viewNotificationBuildStartItem
        ]


viewNotificationsPanelHeading : Element msg
viewNotificationsPanelHeading =
    row
        [ width fill
        , Font.color Palette.neutral7
        , Font.extraLight
        , Font.size 17
        ]
        [ text "Recent activity" ]


viewNotificationCommitItem : Element msg
viewNotificationCommitItem =
    row
        [ Border.width 1
        , Border.color Palette.primary4
        , Font.color Palette.neutral4
        , Font.light
        , Border.dashed
        , Border.rounded 5
        , padding 10
        , width fill
        , mouseOver [ Background.color Palette.primary3, Font.color Palette.neutral5 ]
        ]
        [ el
            [ width (px 25)
            , height (px 25)
            , Background.image "https://i.imgur.com/4vEcq8U.png"
            , Border.rounded 10
            ]
            none
        , el [ width fill ]
            (paragraph
                [ Font.size 15
                , alignLeft
                , paddingXY 10 0
                ]
                [ el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text "Eddy Lane")
                , el [ alignLeft, Font.color Palette.neutral5 ] (text " created commit ")
                , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text "b3e8a32")
                , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text "8 hours ago")
                ]
            )
        ]


viewNotificationBuildStartItem : Element msg
viewNotificationBuildStartItem =
    row
        [ Border.width 1
        , Border.color Palette.primary4
        , Font.color Palette.neutral4
        , Font.light
        , Border.dashed
        , Border.rounded 5
        , padding 10
        , width fill
        , mouseOver [ Background.color Palette.primary3, Font.color Palette.neutral5 ]
        ]
        [ el
            [ width (px 25)
            , height (px 25)
            , Background.image "https://i.imgur.com/4vEcq8U.png"
            , Border.rounded 10
            ]
            none
        , el [ width fill ]
            (paragraph
                [ Font.size 15
                , alignLeft
                , paddingXY 10 0
                ]
                [ el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text "Eddy Lane")
                , el [ alignLeft, Font.color Palette.neutral5 ] (text " started build ")
                , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text "fdsfds")
                , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text "9 hours ago")
                ]
            )
        ]
