module Component.BranchFilter exposing (Context, Config, DropdownState, initialDropdownState, view, subscriptions)

{- A stateless BranchFilter component. -}
-- EXTERNAL --

import Html exposing (..)
import Html.Events exposing (onWithOptions, onInput)
import Html.Attributes exposing (..)
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Json.Decode as Decode


-- INTERNAL --

import Util exposing ((=>))
import Data.Branch as Branch exposing (Branch)


-- MODEL --


type alias Context =
    { branches : List Branch
    , dropdownState : Dropdown.State
    , filterTerm : String
    }


type alias Config msg =
    { dropdownMsg : Dropdown.State -> msg
    , termMsg : String -> msg
    , noOpMsg : msg
    , selectBranchMsg : Maybe Branch -> msg
    }


type alias DropdownState =
    Dropdown.State


initialDropdownState : DropdownState
initialDropdownState =
    Dropdown.initialState



-- SUBSCRIPTIONS --


subscriptions : Config msg -> Context -> Sub msg
subscriptions { dropdownMsg } { dropdownState } =
    Dropdown.subscriptions dropdownState dropdownMsg



-- VIEW --


view : Config msg -> Context -> Html msg
view config context =
    div []
        [ Dropdown.dropdown
            context.dropdownState
            { options = [ Dropdown.menuAttrs [ onClick (config.noOpMsg) ] ]
            , toggleMsg = config.dropdownMsg
            , toggleButton =
                Dropdown.toggle
                    [ Button.outlineSecondary ]
                    [ i [ class "fa fa-code-fork" ] []
                    , text " All branches"
                    ]
            , items = viewDropdownItems config context
            }
        ]


viewDropdownItems : Config msg -> Context -> List (Dropdown.DropdownItem msg)
viewDropdownItems config context =
    let
        filterForm =
            Dropdown.customItem (viewForm config.termMsg context)

        branchItems =
            viewBranchItems config context
    in
        filterForm :: Dropdown.divider :: branchItems


viewBranchItems : Config msg -> Context -> List (Dropdown.DropdownItem msg)
viewBranchItems config { branches, filterTerm } =
    branches
        |> List.filter (branchFilter filterTerm)
        |> List.sortBy (.name >> Just >> Branch.nameToString)
        |> List.map (viewBranchItem config.selectBranchMsg)


branchFilter : String -> Branch -> Bool
branchFilter filterTerm { name, active } =
    active && String.contains filterTerm (Branch.nameToString (Just name))


viewBranchItem : (Maybe Branch -> msg) -> Branch -> Dropdown.DropdownItem msg
viewBranchItem selectMsg branch =
    Dropdown.buttonItem
        [ onClick (selectMsg <| Just branch) ]
        [ text (Just branch.name |> Branch.nameToString) ]


viewForm : (String -> msg) -> Context -> Html msg
viewForm msg { filterTerm } =
    Form.form [ class "px-2 py-1", style [ "width" => "400px" ] ]
        [ Form.group []
            [ Input.email
                [ Input.id "filter-branch-input"
                , Input.placeholder "Filter branches"
                , Input.attrs [ onInput msg ]
                ]
            ]
        ]



-- helper to cancel click anywhere


onClick : msg -> Attribute msg
onClick message =
    onWithOptions
        "click"
        { stopPropagation = True
        , preventDefault = False
        }
        (Decode.succeed message)
