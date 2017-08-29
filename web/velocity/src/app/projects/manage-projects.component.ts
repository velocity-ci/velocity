import { Component, OnInit } from '@angular/core';
import { Project } from "../models/project.model";
import { Validators, FormControl, FormGroup } from "@angular/forms";
import { ProjectRepository } from "./project.repository";

@Component({
    selector: 'app-manage-projects',
    templateUrl: './manage-projects.component.html'
})
export class ManageProjectsComponent implements OnInit {

    projectForm: FormGroup;
    project = new Project();

    submitting = false;

    constructor(private projectRepository: ProjectRepository) {}

    ngOnInit(): void {
        this.projectForm = new FormGroup({
            'name': new FormControl(this.project.name, [
                Validators.required,
                Validators.minLength(3)
            ]),
            'repository': new FormControl(this.project.repository, [
                Validators.required,
                Validators.minLength(8)
            ]),
            'key': new FormControl(this.project.key, [
                Validators.required,
                Validators.minLength(8)
            ])
        })
    }

    submit() {
        this.submitting = true;
        this.projectRepository.createProject(this.project).then((res) => {
          console.log(res);
        });
        // this.authService.authenticate(this.loginUser).then(() => {

        // }).catch((response) => {
        //     console.log(response);
        //     switch(response.status) {
        //         case 401:
        //             this.loginForm.controls.username.setErrors({
        //                 'username': "invalid"
        //             });
        //             this.loginForm.controls.password.setErrors({
        //                 'password': "invalid"
        //             });
        //     }
        //     this.submitting = false;
        // });
    }
}
