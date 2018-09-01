/*
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS HEADER.
 *
 * Copyright (c) 2018 Oracle and/or its affiliates. All rights reserved.
 *
 * The contents of this file are subject to the terms of either the GNU
 * General Public License Version 2 only ("GPL") or the Common Development
 * and Distribution License("CDDL") (collectively, the "License").  You
 * may not use this file except in compliance with the License.  You can
 * obtain a copy of the License at
 * https://glassfish.dev.java.net/public/CDDL+GPL_1_1.html
 * or packager/legal/LICENSE.txt.  See the License for the specific
 * language governing permissions and limitations under the License.
 *
 * When distributing the software, include this License Header Notice in each
 * file and include the License file at packager/legal/LICENSE.txt.
 *
 * GPL Classpath Exception:
 * Oracle designates this particular file as subject to the "Classpath"
 * exception as provided by Oracle in the GPL Version 2 section of the License
 * file that accompanied this code.
 *
 * Modifications:
 * If applicable, add the following below the License Header, with the fields
 * enclosed by brackets [] replaced by your own identifying information:
 * "Portions Copyright [year] [name of copyright owner]"
 *
 * Contributor(s):
 * If you wish your version of this file to be governed by only the CDDL or
 * only the GPL Version 2, indicate your decision by adding "[Contributor]
 * elects to include this software in this distribution under the [CDDL or GPL
 * Version 2] license."  If you don't indicate a single choice of license, a
 * recipient has the option to distribute your version of this file under
 * either the CDDL, the GPL Version 2 or to extend the choice of license to
 * its licensees as provided above.  However, if you add GPL Version 2 code
 * and therefore, elected the GPL Version 2 license, then the option applies
 * only if the new code is made subject to such option by the copyright
 * holder.
 */

package example

import (
	"fmt"
	"github.com/jwells131313/dargo/ioc"
)

// MusicService can play a scale!
type MusicService interface {
	PlayCScale() string
}

// NoteService can play a note!
type NoteService interface {
	PlayNote() string
}

type musicPlayer struct {
	c, d, e, f, g, a, b NoteService
}

type noteGenerator struct {
	note string
}

func newMusicPlayer(c, d, e, f, g, a, b NoteService) MusicService {
	return &musicPlayer{
		c: c,
		d: d,
		e: e,
		f: f,
		g: g,
		a: a,
		b: b,
	}
}

func newNoteService(note string) NoteService {
	return &noteGenerator{
		note: note,
	}
}

func (player *musicPlayer) PlayCScale() string {
	return fmt.Sprintf("<<<%s%s%s%s%s%s%s>>>",
		player.c.PlayNote(),
		player.d.PlayNote(),
		player.e.PlayNote(),
		player.f.PlayNote(),
		player.g.PlayNote(),
		player.a.PlayNote(),
		player.b.PlayNote())
}

func (note *noteGenerator) PlayNote() string {
	return note.note
}

const (
	// NoteServiceName is the name of the note service.  Note services also have qualifiers
	// for the specific note they play
	NoteServiceName = "NoteServiceName"
	// MusicServiceName is the name of the music service
	MusicServiceName = "MusicServiceName"
)

func start() (ioc.ServiceLocator, error) {
	return ioc.CreateAndBind("OracleMusicPlayer", func(binder ioc.Binder) error {
		// bind the "A" note
		binder.BindWithCreator(NoteServiceName, func(locator ioc.ServiceLocator, key ioc.Descriptor) (interface{}, error) {
			return newNoteService("a"), nil
			// Give the service the NoteServiceName and qualify it with the note
		}).QualifiedBy("a")

		// bind the "B" note
		binder.BindWithCreator(NoteServiceName, func(locator ioc.ServiceLocator, key ioc.Descriptor) (interface{}, error) {
			return newNoteService("b"), nil
			// Give the service the NoteServiceName and qualify it with the note
		}).QualifiedBy("b")

		// I got tired of doing it one-by-one, so I'm doing the rest in a loop

		remainingNotes := []string{"c", "d", "e", "f", "g"}
		for _, remainingNote := range remainingNotes {
			// allocate new storage for each note variable
			note := remainingNote

			// BindWithCreator all the other notes in this loop
			binder.BindWithCreator(NoteServiceName, func(locator ioc.ServiceLocator, key ioc.Descriptor) (interface{}, error) {
				// Creates a new note service with the given note
				return newNoteService(note), nil

				// Give the service the NoteServiceName and qualify it with the note
			}).QualifiedBy(remainingNote)
		}

		return nil
	})
}

func bindPlayer(locator ioc.ServiceLocator) error {
	return ioc.BindIntoLocator(locator, func(binder ioc.Binder) error {
		binder.BindWithCreator(MusicServiceName, createMusicService)
		return nil
	})
}

func createMusicService(locator ioc.ServiceLocator, key ioc.Descriptor) (interface{}, error) {
	notes := []string{"a", "b", "c", "d", "e", "f", "g"}
	serviceMap := make(map[string]NoteService)

	for _, note := range notes {
		noteService, err := locator.GetDService(NoteServiceName, note)
		if err != nil {
			return nil, err
		}

		serviceMap[note] = noteService.(NoteService)
	}

	musicPlayer := newMusicPlayer(
		serviceMap["c"],
		serviceMap["d"],
		serviceMap["e"],
		serviceMap["f"],
		serviceMap["g"],
		serviceMap["a"],
		serviceMap["b"])

	return musicPlayer, nil
}
