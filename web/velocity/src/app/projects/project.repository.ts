import { Injectable } from '@angular/core';
import 'rxjs/add/operator/toPromise';
import { APIService } from "../api.service";
import { Project } from "../models/project.model";

@Injectable()
export class ProjectRepository {

    private projects: Map<string, Project>

    constructor(
        private apiService: APIService,
    ) {
        this.projects = new Map();
    }

    public getAllProjects(): Project[] {
        return Array.from(this.projects.values())
    }

    public createProject(p: Project) {
        return this.apiService.post('/v1/projects', p)
            .then(res => {
                const p = res.data.json() as Project;
                this.projects.set(p.id, p);
            })
        ;
    }

    public updateProject(p: Project) {

    }

    public deleteProject() {

    }
}
