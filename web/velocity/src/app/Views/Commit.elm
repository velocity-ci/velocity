module Views.Commit exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Branch as Branch exposing (Branch)
import Page.Helpers exposing (formatTime, formatDateTime)


commitTimeInformation : Commit -> Html msg
commitTimeInformation commit =
    small [] [ strong [] [ text commit.author ], text (" commited at " ++ formatTime commit.date) ]


infoPanel : Commit -> Html msg
infoPanel commit =
    small []
        [ strong [] [ text (Commit.truncateHash commit.hash) ]
        , text " by "
        , strong [] [ text commit.author ]
        , text " at "
        , strong [] [ text <| formatDateTime commit.date ]
        ]


branchList : Commit -> Html msg
branchList commit =
    ul [ class "mb-0 list-inline" ] <| List.map branch commit.branches


branch : Branch.Name -> Html msg
branch branch =
    li [ class "list-inline-item" ]
        [ span [ class "badge badge-secondary", style [ ( "white-space", "pre-line" ) ] ]
            [ i [ class "fa fa-code-fork" ] []
            , text (" " ++ (Branch.nameToString (Just branch)))
            ]
        ]


truncateCommitMessage : Commit -> String
truncateCommitMessage commit =
    if (String.length commit.message) > 48 then
        (String.slice 0 44 commit.message) ++ "..."
    else
        commit.message
