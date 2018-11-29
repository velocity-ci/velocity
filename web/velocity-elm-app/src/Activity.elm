module Activity exposing (Log, ViewConfiguration, init, projectAdded, unreadAmount, view)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Palette
import Project exposing (Project)
import Project.Id as Project
import Set exposing (Set)



-- TYPES


type ActivityItem
    = ActivityItem Int Activity


type Activity
    = Project CategoryProject


type CategoryProject
    = ProjectAdded Project.Id



-- LOG


type Log
    = Log Internals


type alias Internals =
    { currentId : Int
    , activities : List ActivityItem
    , seenActivities : Set Int
    }


init : Log
init =
    Log <|
        { currentId = 0
        , activities = []
        , seenActivities = Set.empty
        }


projectAdded : Project.Id -> Log -> Log
projectAdded id (Log internals) =
    let
        incrementedId =
            internals.currentId + 1

        activityItem =
            ActivityItem incrementedId <|
                Project <|
                    ProjectAdded id
    in
    Log <|
        { internals | activities = activityItem :: internals.activities }



---- INFO


unreadAmount : Log -> Int
unreadAmount (Log { activities, seenActivities }) =
    List.foldl
        (\(ActivityItem id _) acc ->
            if not <| Set.member id seenActivities then
                acc + 1

            else
                acc
        )
        0
        activities



---- VIEW


type alias ViewConfiguration =
    { activities : Log
    , projects : List Project
    }


view : ViewConfiguration -> Element msg
view config =
    let
        (Log { activities }) =
            config.activities
    in
    column
        [ Background.color Palette.primary2
        , width fill
        , height fill
        , paddingEach { top = 80, bottom = 90, left = 20, right = 20 }
        , spacing 10
        ]
        (viewNotificationsPanelHeading :: viewNotifications config)


viewNotificationsPanelHeading : Element msg
viewNotificationsPanelHeading =
    row
        [ width fill
        , Font.color Palette.neutral7
        , Font.extraLight
        , Font.size 17
        ]
        [ text "Recent activity" ]


viewNotifications : ViewConfiguration -> List (Element msg)
viewNotifications config =
    let
        (Log { activities }) =
            config.activities
    in
    List.map (viewNotification config.projects) activities


viewNotification : List Project -> ActivityItem -> Element msg
viewNotification projects activity =
    case activity of
        ActivityItem _ (Project category) ->
            viewProjectNotification projects category


viewProjectNotification : List Project -> CategoryProject -> Element msg
viewProjectNotification projects category =
    let
        maybeProject =
            projectFromCategory projects category
    in
    case maybeProject of
        Just project ->
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
                [ el [ width (px 40), height (px 40) ] <| Project.thumbnail project
                , el [ width fill ]
                    (paragraph
                        [ Font.size 15
                        , alignLeft
                        , paddingXY 10 0
                        ]
                        [ el [ alignLeft, Font.color Palette.neutral5 ] (text "Project ")
                        , el [ Font.heavy, alignLeft, Font.color Palette.neutral6 ] (text <| Project.name project)
                        , el [ alignLeft, Font.color Palette.neutral5 ] (text " created ")
                        , el [ Font.extraLight, Font.size 13, Font.color Palette.neutral5, alignLeft ] (text "8 hours ago")
                        ]
                    )
                ]

        Nothing ->
            none


projectFromCategory : List Project -> CategoryProject -> Maybe Project
projectFromCategory projects category =
    case category of
        ProjectAdded projectId ->
            Project.findProject projects projectId
