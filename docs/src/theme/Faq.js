import React, { useState } from 'react'
import cn from 'classnames'
import styles from './faq.module.css'
import tags from './faq_tags.json'
import questions from './faq_questions.json'

const TagButton = ({ isSelected, children, toggleSelected  }) => (
    <button
          className={cn({ [styles.selected]: isSelected })}
          onClick={toggleSelected}
        >
          {children}
        </button>
)

const Faq = () => {
  const [selectedTags, setSelectedTags] = useState(tags)
  const displayFunc = (tags) => { 
    for (var i = 0; i < tags.length; i++) {
        if (selectedTags.find((t) => t === tags[i]) ) {
            return 'block'
        }
    }
    return 'none'
  } 
  return (
    <>
      {tags.map((tag) => (
        <TagButton
        key={tag}
        isSelected={selectedTags.find((t) => t === tag)}
        toggleSelected={() => {
            if (selectedTags.find((t) => t === tag)) {
                setSelectedTags(selectedTags.filter((t) => t !== tag))
            } else {
                setSelectedTags([...selectedTags, tag])
            }
        }}
        >
            #{tag}
        </TagButton>
      ))}

      {questions.map((question) => (
          
          <div style={{display: 
            displayFunc(question.tags)
          }}>
              M: {question.tags}<br/>
              Q: {question.q}<br/>
              A: {question.a}<br/>
          </div>
      ))}

    </>
  )
}

export default Faq
