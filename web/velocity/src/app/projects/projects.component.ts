import { Component, OnInit } from '@angular/core';
import { Project } from "../models/project.model";
import { Validators, FormControl, FormGroup } from "@angular/forms";
import { ProjectRepository } from "./project.repository";

@Component({
    selector: 'app-projects',
    templateUrl: './projects.component.html'
})
export class ProjectsComponent implements OnInit {

    projectForm: FormGroup;
    project = new Project();

    submitting = false;
    createProjectCollapsed = true;

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
        });
        this.projectRepository.refreshProjects();
    }

    submit() {
        this.submitting = true;
        this.projectRepository.createProject(this.project).then(res => {
          this.submitting = false;
          this.projectForm.reset();
          this.createProjectCollapsed = true;
        });
    }
}
